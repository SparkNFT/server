package worker

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/SparkNFT/key_server/abi"
	"github.com/SparkNFT/key_server/chain"
	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/model"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"xorm.io/xorm"
)

var (
	scanLock map[string]*sync.Mutex
)

func CheckBlockScannerConfig(chainName string) {
	if config.C.Chain[chainName].ContractAddress == "" || config.C.Chain[chainName].RPCUrl == "" {
		panic(fmt.Sprintf(
			"Chain config invalid. ContractAddress: %s, RPC URL: %s",
			config.C.Chain[chainName].ContractAddress,
			config.C.Chain[chainName].RPCUrl,
		))
	}

	if config.C.Chain[chainName].BlockHeight == 0 {
		panic("BlockHeight not set.")
	}
}

func BlockScannerWorker(chainName string) {
	blockHeight := checkBlockHeight(chainName)
	l := log.WithFields(log.Fields{"chain": chainName, "worker": "BlockScannerWorker"})

	contract, client, err := chain.Init(chainName)
	if err != nil {
		panic(xerrors.Errorf("error when initializing worker: %w", err))
	}
	if scanLock == nil {
		scanLock = make(map[string]*sync.Mutex, 0)
	}
	lock := new(sync.Mutex)
	scanLock[chainName] = lock

	for {
		lock.Lock()

		if err := fetch_block(chainName, contract, client, blockHeight); err != nil {
			l.WithFields(logrus.Fields{"chain": chainName, "height": blockHeight}).Warnf("Block fetch failed: %s", err.Error())
			lock.Unlock()
			time.Sleep(config.C.Chain[chainName].FailSleepSeconds * time.Second)
			continue
		}

		blockHeight += 1
		lock.Unlock()
		time.Sleep(config.C.Chain[chainName].SleepSeconds * time.Second)
	}
}

// checkBlockHeight returns next block height should be fetched
func checkBlockHeight(chainName string) (block_height uint64) {
	l := logrus.WithFields(log.Fields{"chain": chainName, "worker": "check_block_height"})
	config_block_height := config.C.Chain[chainName].BlockHeight

	found_block, err := model.BlockLogFindFirst(chainName)
	if err != nil {
		l.Warnf("error when fetching latest block height: %s", err.Error())
		l.Warnf("Using config-file specified height")
	}

	found_block_height := uint64(0)
	if found_block != nil {
		found_block_height = found_block.BlockHeight
		found_block_height += 1
	}

	if config_block_height > found_block_height {
		block_height = config_block_height
	} else {
		block_height = found_block_height
	}
	if block_height == uint64(0) {
		panic(xerrors.Errorf("Block height: both config file and DB fetching failed."))
	}
	return block_height
}

// fetch_block fetches specific height of a block and saves all
// related events in DB.
func fetch_block(chainName string, contract *abi.SparkLink, client *ethclient.Client, block_height uint64) (err error) {
	l := log.WithFields(log.Fields{"chain": chainName, "worker": "fetch_block", "height": block_height})
	newest_block, err := client.BlockNumber(context.Background())
	if err != nil {
		return xerrors.Errorf("error when fetching newest block number: %w", err)
	}
	wait_block_count := uint64(config.C.Chain[chainName].BlockConfirmCount)
	if newest_block < (block_height + wait_block_count) {
		return xerrors.Errorf("Slow down. Newest: %d, current: %d, should wait: %d", newest_block, block_height, wait_block_count)
	}

	session := model.Engine.NewSession()
	defer session.Close()

	err = model.BlockLogStart(session, chainName, block_height)
	if err != nil {
		session.Rollback()
		if strings.Contains(err.Error(), "duplicate key value") {
			l.WithField("height", block_height).Debug("Using existed BlockLog record")
		} else {
			return xerrors.Errorf("error when starting a blocklog: %w", err)
		}

	}

	logs, err := chain.GetAllLogsOf(client, chainName, big.NewInt(int64(block_height)))
	if err != nil {
		session.Rollback()
		return xerrors.Errorf("error when fetching block: %w", err)
	}

	l.WithFields(log.Fields{"count": len(logs)}).Info("Log fetched.")
	events, err := create_events(contract, session, chainName, logs)
	if err != nil {
		session.Rollback()
		return xerrors.Errorf("%w", err)
	}
	err = session.Commit()
	if err != nil {
		return xerrors.Errorf("error when commiting changes: %w", err)
	}

	// create NFT from events
	err = create_nfts(contract, session, chainName, events)
	if err != nil {
		return xerrors.Errorf("error when creating NFT: %w", err)
	}
	err = update_nfts(contract, session, chainName, events)
	if err != nil {
		return xerrors.Errorf("error when updating NFT: %w", err)
	}
	err = session.Commit()
	if err != nil {
		return xerrors.Errorf("error when commiting changes: %w", err)
	}

	// If all set,
	err = model.BlockLogFinish(session, chainName, block_height)
	if err != nil {
		return xerrors.Errorf("error when finishing a block: %w", err)
	}

	err = session.Commit()
	if err != nil {
		return xerrors.Errorf("error when commiting changes: %w", err)
	}

	// Clean old BlockLog
	err = model.BlockLogClean(chainName)
	if err != nil {
		return xerrors.Errorf("error when cleaning BlockLog: %w", err)
	}

	return nil
}

// create_events filter and save all logs as events in DB.
func create_events(contract *abi.SparkLink, session *xorm.Session, chainName string, logs []types.Log) (result []*model.Event, err error) {
	publish_events, err := chain.FilterEventPublish(contract, logs)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	publish, err := model.CreateFromBlockEventPublish(session, chainName, publish_events)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	transfer_events, err := chain.FilterEventTransfer(contract, logs)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	transfer, err := model.CreateFromBlockEventTransfer(session, chainName, transfer_events)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	result = make([]*model.Event, 0, len(logs))
	result = append(result, publish...)
	result = append(result, transfer...)
	return result, nil
}

// create_nfts create NFT model records
func create_nfts(contract *abi.SparkLink, session *xorm.Session, chainName string, events []*model.Event) (err error) {
	if len(events) == 0 {
		return nil
	}

	l := log.WithFields(log.Fields{"chain": chainName, "worker": "create_nfts"})
	nfts := make([]*model.NFT, 0)
	existed_parent_nft_ids := make([]uint64, 0)
	for _, event := range events {
		if !event.IsMint() {
			continue
		}
		parent, err := chain.GetParentOf(contract, event.NFTId) // parent == 0 : Root NFT
		if err != nil {
			return xerrors.Errorf("error when getting parent of %d: %w", event.NFTId, err)
		}
		max_shill_count, err := contract.GetShillTimesByNFTId(nil, event.NFTId)
		if err != nil {
			return xerrors.Errorf("error when getting remain shill count of %d: %w", event.NFTId, err)
		}

		token_addr, err := contract.GetTokenAddrByNFTId(nil, event.NFTId)
		if err != nil {
			return xerrors.Errorf("error when getting TokenAddr of %d: %w", event.NFTId, err)
		}

		nft := &model.NFT{
			Chain:         chainName,
			NFTID:         event.NFTId,
			Parent:        parent,
			ShillCount:    0,
			MaxShillCount: max_shill_count,
			Owner:         event.To,
			TokenAddr:     token_addr.Hex(),
		}
		if parent != uint64(0) {
			existed_parent_nft_ids = append(existed_parent_nft_ids, parent)
		}
		nfts = append(nfts, nft)
	}

	if len(nfts) != 0 {
		affected, err := model.Engine.Insert(&nfts)
		if err != nil {
			return xerrors.Errorf("error when inserting NFT record: %w", err)
		}
		if int(affected) != len(nfts) {
			return xerrors.Errorf("error when Inserting NFT Record: records inserted mismatch total length: %d - %d", affected, len(nfts))
		}
	} else {
		l.Debugf("No NFT created.")
	}

	if len(existed_parent_nft_ids) > 0 {
		err := increase_shill_count(session, chainName, existed_parent_nft_ids)
		if err != nil {
			return xerrors.Errorf("%w", err)
		}
	}

	return nil
}

// increase_shill_count increases 1 shill count for every given nft ids
func increase_shill_count(session *xorm.Session, chainName string, nft_ids []uint64) error {
	l := log.WithFields(log.Fields{"worker": "increase_shill_count"})
	for _, nft_id := range nft_ids {
		nft, err := model.FindNFT(chainName, nft_id)
		if err != nil {
			return xerrors.Errorf("error when finding NFT ID %d: %w", nft_id, err)
		}

		nft.ShillCount += 1
		l.WithFields(log.Fields{"NFTID": nft.NFTID, "ShillCount": nft.ShillCount}).Infof("Updating NFT shill count")
		_, err = session.Cols("shill_count").ID(nft.Id).Update(nft)
		if err != nil {
			return xerrors.Errorf("error when updating NFT %d: %w", nft.NFTID, err)
		}
	}

	return nil
}

// update_nfts update existed NFT records
func update_nfts(contract *abi.SparkLink, session *xorm.Session, chainName string, events []*model.Event) (err error) {
	if len(events) == 0 {
		return nil
	}
	l := log.WithFields(log.Fields{"worker": "update_nfts"})
	for _, event := range events {
		if event.IsMint() {
			continue
		}

		nft, err := model.FindNFT(chainName, event.NFTId)
		if err != nil {
			return xerrors.Errorf("error when update NFT %d: %w", event.NFTId, err)
		}

		// Update NFT fields
		l.WithFields(log.Fields{"NFTID": nft.NFTID, "Owner": event.To}).Infof("Updating NFT owner")
		nft.Owner = event.To

		_, err = session.Cols("owner").ID(nft.Id).Update(nft)
		if err != nil {
			return xerrors.Errorf("error when updating NFT %d: %w", nft.NFTID, err)
		}
	}
	return nil
}

package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/SparkNFT/key_server/chain"
	"github.com/SparkNFT/key_server/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

var flagConfigPath = flag.String("config", "./config.json", "Location of config.json")
var flagFirstSellPrice = flag.Int64("first", int64(100), "First sell price")
var flagShillTimes = flag.Uint64("shill", 10, "Max shill times")
var flagRoyaltyFee = flag.Uint("royalty", 10, "Royalty fee")
var flagIpfsHash = flag.String("ipfs", "7ba461e1c8994e110a1f371af1a9ed01490344e582c65ec1cd139615fb7b3bfd", "IPFS Hash (64 hex digits, must NOT start with 0x)")

func wait_tx(client *ethclient.Client, tx *types.Transaction) (finished_tx *types.Transaction) {
	time.Sleep(time.Second * 3) // Make sure transaction is broadcasted
	for {
		finished_tx, pending, err := client.TransactionByHash(context.Background(), tx.Hash())
		if err != nil {
			panic(xerrors.Errorf("Error during waiting tx: %s", err.Error()))
		}
		if pending {
			logrus.WithField("TX Hash", tx.Hash()).Debug("Waiting...")
			time.Sleep(time.Second * 3)
		} else {
			logrus.WithField("TX Hash", tx.Hash()).Debug("Finished")
			return finished_tx
		}
	}
}

func main() {
	chainName := "ethereum"
	flag.Parse()
	config.ConfigPath = *flagConfigPath
	config.Init()
	logrus.SetLevel(logrus.DebugLevel)

	contract, client, err := chain.Init(chainName)
	if err != nil {
		panic(xerrors.Errorf("error when initializing contract: %w", err))
	}

	txops, err := chain.GenerateTransactOps(client, config.C.Chain[chainName].OperatorAccountPrivateKey)
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	logrus.Infof("Now using account: %s", txops.From.String())

	var ipfs_hash_bytes [32]byte
	logrus.Debugf("IPFS Hash: 0x%s", *flagIpfsHash)
	if len(*flagIpfsHash) != 64 {
		panic(fmt.Sprintf("IPFS Hash length error, current length: %d", len(*flagIpfsHash)))
	}
	ipfs_hash_slice, err := hex.DecodeString(*flagIpfsHash)
	if err != nil {
		panic(xerrors.Errorf("error when decoding IPFS hash: %w", err))
	}
	copy(ipfs_hash_bytes[:], ipfs_hash_slice)

	tx, err := contract.Publish(
		txops,
		big.NewInt(*flagFirstSellPrice),
		uint8(*flagRoyaltyFee),
		uint16(*flagShillTimes),
		ipfs_hash_bytes,
		common.HexToAddress("0x0"),
		true,
		true,
		true,
	)
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	logrus.WithField("TX Hash", tx.Hash()).Info("Sent")
	wait_tx(client, tx)
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}

	logrus.WithField("Block", receipt.BlockNumber.Uint64()).Debug("")
	logs, err := chain.GetAllLogsOf(client, chainName, receipt.BlockNumber)
	if err != nil {
		panic(xerrors.Errorf("error when fetching block events: %w", err))
	}
	for index, log := range logs {
		logrus.WithField("Block", receipt.BlockNumber.Uint64()).Debugf("Log #%d : %+v\n", index, log)
	}

	transfer_events, err := chain.FilterEventTransfer(contract, logs)
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	for _, event := range transfer_events {
		logrus.WithField("NFT_ID", event.TokenId.Text(16)).Info("NFT generated and transfered")
	}

	os.Exit(0)
}

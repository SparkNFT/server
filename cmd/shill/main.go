package main

import (
	"context"
	"flag"
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
var (
	flagConfig = flag.String("config", "./config.json", "config.json file path")
	flagNftId = flag.String("nftid", "0000000000000000000000000000001000000000000000000000000000000001", "NFT ID (Hexstring without '0x')")
)

func wait_tx(client *ethclient.Client, tx *types.Transaction) (receipt *types.Receipt) {
	time.Sleep(time.Second * 3) // Make sure transaction is broadcasted
	for {
		_, pending, err := client.TransactionByHash(context.Background(), tx.Hash())
		if err != nil {
			panic(xerrors.Errorf("Error during waiting tx: %s", err.Error()))
		}
		if pending {
			logrus.Debugf("Waiting for tx %s to finish...", tx.Hash())
			time.Sleep(time.Second * 3)
		} else {
			logrus.Debugf("Tx %s finished.", tx.Hash())
			break
		}
	}

	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	return receipt
}

func main() {
	flag.Parse()
	config.ConfigPath = *flagConfig
	config.Init()
	logrus.SetLevel(logrus.DebugLevel)
	chainName := "ethereum"

	root_nft_id := big.NewInt(0)
	_, success := root_nft_id.SetString(*flagNftId, 16)
	if !success {
		panic("Parse NFT ID failed")
	}
	logrus.Debugf("Shilling NFT: %s (%+v)", common.BigToHash(root_nft_id).Hex(), root_nft_id)

	contract, client, err := chain.Init("ethereum")
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}

	result, err := contract.IsEditionExisting(nil, root_nft_id.Uint64())
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	if result == false {
		panic(xerrors.Errorf("IsEditionExist() returned false for NFT ID: %+v", root_nft_id))
	}

	auth, err := chain.GenerateTransactOps(client, config.C.Chain["ethereum"].OperatorAccountPrivateKey)
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	logrus.Debugf("Now using %s\n", auth.From.Hex())
	value, err := contract.GetShillPriceByNFTId(nil, root_nft_id.Uint64())
	if err != nil {
		panic(xerrors.Errorf("error when getting shill price: %w", err))
	}
	logrus.Infof("Shill price: %v", value)
	auth.Value = value

	tx, err := contract.AcceptShill(auth, root_nft_id.Uint64())
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	receipt := wait_tx(client, tx)
	if len(receipt.Logs) == 0 {
		panic(xerrors.Errorf("Receipt Logs length is 0. Maybe this tx is failed."))
	}

	logs, _ := chain.GetAllLogsOf(client, chainName, receipt.BlockNumber)
	transfer_events, err := chain.FilterEventTransfer(contract, logs)
	if err != nil {
		panic(xerrors.Errorf("%w", err))
	}
	for _, event := range transfer_events {
		logrus.WithField("NFT_ID", event.TokenId.Text(16)).Info("NFT generated")
	}

	os.Exit(0)
}

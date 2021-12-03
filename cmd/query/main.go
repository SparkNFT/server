package main

import (
	"flag"
	"fmt"
	"math/big"

	"github.com/SparkNFT/key_server/chain"
	"github.com/SparkNFT/key_server/config"
)

var (
	flagConfig = flag.String("config", "./config/config.json", "config.json file path")
	flagChain = flag.String("chain", "ethereum", "target chain")
	flagNftId = flag.Uint64("nftid", uint64(0), "target NFT ID (decimal)")
)

func main() {
	flag.Parse()

	config.ConfigPath = *flagConfig
	config.Init()

	contract, _, err := chain.Init(*flagChain)
	if err != nil {
		panic(err.Error())
	}

	remainTimes, err := contract.GetRemainShillTimesByNFTId(nil, *flagNftId)
	if err != nil {
		panic(err.Error())
	}

	totalTimes, err := contract.GetShillTimesByNFTId(nil, *flagNftId)
	if err != nil {
		panic(err)
	}

	tokenAddr, err := contract.GetTokenAddrByNFTId(nil, *flagNftId)
	if err != nil {
		panic(err)
	}

	owner, err := contract.OwnerOf(nil, new(big.Int).SetUint64(*flagNftId) )
	if err != nil {
		panic(err)
	}

	parent, err := contract.GetFatherByNFTId(nil, *flagNftId)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Owner: %s\n", owner.String())
	fmt.Printf("Token Address: %s\n", tokenAddr.String())
	fmt.Printf("Shill times: %d / %d\n", (totalTimes - remainTimes), totalTimes)
	fmt.Printf("Parent: %d\n", parent)
}

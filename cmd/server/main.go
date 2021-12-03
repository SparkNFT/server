package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/controller"
	"github.com/SparkNFT/key_server/model"
	"github.com/SparkNFT/key_server/worker"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

const (
	ENVIRONMENT    = "development"
	LISTEN_ADDRESS = "0.0.0.0:3000"
)

var flagConfig = flag.String("config", "./config.json", "config.json path")
var flagDebug = flag.Bool("debug", false, "Enable debug-level log")
var flagChains = flag.String("chains", "", "All enabled chains, separeted by comma. If not given, all available chain in config file will be enabled.")

func main() {
	flag.Parse()
	if *flagDebug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	config.ConfigPath = *flagConfig
	config.Init()

	model.Init()
	controller.Init()

	enableChains()

	for chainName, chainConfig := range config.C.Chain {
		if chainConfig.Enabled == true {
			worker.CheckBlockScannerConfig(chainName)
			go worker.BlockScannerWorker(chainName)
		}
	}

	err := controller.Engine.Run(LISTEN_ADDRESS)
	if err != nil {
		panic(xerrors.Errorf("error when opening controller: %w", err))
	}

	fmt.Printf("Server listening at %s", LISTEN_ADDRESS)
}

func enableChains() {
	if (*flagChains == "") { // Enable all chain in config
		for _, chainConfig := range config.C.Chain {
			chainConfig.Enabled = true
		}
		return
	}

	enabledChains := strings.Split(*flagChains, ",")
	for _, enabledChain := range enabledChains {
		chainConfig, ok := config.C.Chain[enabledChain]
		if !ok {
			panic(fmt.Sprintf("Chain '%s' not found in config file.", enabledChain))
		}
		chainConfig.Enabled = true
	}
}

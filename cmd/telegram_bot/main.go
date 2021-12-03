package main

import (
	"flag"

	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/model"
	"github.com/SparkNFT/key_server/telegram"
	"github.com/sirupsen/logrus"
)

var (
	flag_config = flag.String("config", "./config.json", "config.json path")
)

func main() {
	flag.Parse()
	config.ConfigPath = *flag_config
	config.Init()
	logrus.SetLevel(logrus.DebugLevel)

	model.Init()
	defer model.Engine.Close()

	telegram.Init(false)
	defer telegram.B.Close()

	logrus.WithField("module", "main").Infof("Now listening Telegram messages...")
	// Will block
	telegram.B.Start()
}

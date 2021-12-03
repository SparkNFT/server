package model

import (
	"fmt"

	"github.com/SparkNFT/key_server/config"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
)

var (
	Engine *xorm.Engine
)

// Init initializes ORM - database connection
func Init() {
	if Engine != nil {
		return
	}

	var err error
	Engine, err = xorm.NewEngine("postgres", config.GetDatabaseDSN())
	if err != nil {
		panic(fmt.Sprintf("error during init ORM: %s", err.Error()))
	}

	err = Engine.Sync2(&Key{}, &NFT{}, &BlockLog{}, &Event{})// TODO: finish &TelegramBind{}, &TelegramGroup{}
	if err != nil {
		panic(fmt.Sprintf("error during DB migration: %s", err.Error()))
	}
}

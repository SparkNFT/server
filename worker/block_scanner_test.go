package worker

import (
	"math/rand"
	"testing"
	"time"

	"github.com/SparkNFT/key_server/chain"
	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

const (
	chainName = "ethereum"
)

func before_each(t *testing.T) {
	rand_seed := time.Now().Unix()
	rand.Seed(rand_seed)

	config.ConfigPath = "../config/config.test.json"
	config.Init()
	model.Init()

	db_clean_data()
}

func db_clean_data() {
	model.Engine.Where("1 = 1").Delete(new(model.Key))
	model.Engine.Where("1 = 1").Delete(new(model.BlockLog))
	model.Engine.Where("1 = 1").Delete(new(model.NFT))
	model.Engine.Where("1 = 1").Delete(new(model.Event))
}

func Test_fetch_block(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		before_each(t)
		contract, client, err := chain.Init(chainName)
		assert.Nil(t, err)
		// https://rinkeby.etherscan.io/tx/0xc7a6953f78c0610518888a8d071a87c16f1df210bf8aeee70053da0305f09e81
		height := uint64(9243686)
		block_log := &model.BlockLog{BlockHeight: height}
		found, err := model.Engine.Get(block_log)
		assert.False(t, found)
		assert.Nil(t, err)

		err = fetch_block(chainName, contract, client, height)
		if err != nil {
			t.Logf("%+v", err)
		}
		assert.Nil(t, err)

		publish_events := make([]*model.Event, 0, 1)
		err = model.Engine.Find(
			&publish_events,
			&model.Event{BlockHeight: height, Type: model.EventTypePublish},
		)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(publish_events))

		publish_event := publish_events[0]
		assert.Equal(t, common.HexToAddress("0x0000004215285644116b17436372d569a4ed3a1d").Hex(), publish_event.To)
		assert.Equal(t, common.HexToAddress("0x0").Hex(), publish_event.From)
		assert.Equal(t, model.EventTypePublish, publish_event.Type)
		assert.Equal(t, height, publish_event.BlockHeight)

		transfer_events := make([]*model.Event, 0, 1)
		err = model.Engine.Find(
			&transfer_events,
			&model.Event{BlockHeight: height, Type: model.EventTypeTransfer},
		)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(transfer_events))
		transfer_event := transfer_events[0]

		assert.Equal(t, common.HexToAddress("0x0000004215285644116b17436372d569a4ed3a1d").Hex(), transfer_event.To)
		assert.Equal(t, common.HexToAddress("0x0").Hex(), transfer_event.From)
		assert.Equal(t, model.EventTypeTransfer, transfer_event.Type)
		assert.Equal(t, height, transfer_event.BlockHeight)

		block_log = &model.BlockLog{BlockHeight: height}
		found, err = model.Engine.Get(block_log)
		assert.True(t, found)
		assert.Nil(t, err)
		assert.True(t, block_log.Scanned)
	})
}

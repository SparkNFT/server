package model

import (
	"testing"

	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_IssueId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		before_each(t)
		event := model.Event{
			NFTId: 0x120000000d,
		}
		assert.Equal(t, uint32(0x12), event.IssueId())
	})
}

func Test_EditionId(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		event := model.Event{
			NFTId: 0x120000003d,
		}
		assert.Equal(t, uint32(0x3d), event.EditionId())
	})
}

func Test_IsRoot(t *testing.T) {
	t.Run("when Publish", func(t *testing.T) {
		before_each(t)
		event := model.Event{
			Type: model.EventTypePublish,
			From: config.C.Chain["ethereum"].ContractAddress, // I know it doesn't make sense
		}
		assert.True(t, event.IsRoot())
	})

	t.Run("when From is 0x0", func(t *testing.T) {
		before_each(t)
		event := model.Event{
			Type: model.EventTypeTransfer,
			From: common.HexToAddress("0x0").Hex(),
		}
		assert.True(t, event.IsRoot())
	})

	t.Run("should be false", func(t *testing.T) {
		before_each(t)
		event := model.Event{
			Type: model.EventTypeTransfer,
			From: config.C.Chain["ethereum"].ContractAddress,
		}
		assert.False(t, event.IsRoot())
	})
}

func Test_IsMint(t *testing.T) {
	t.Run("is mint", func(t *testing.T) {
		event := model.Event{
			Type: model.EventTypeTransfer,
			From: "0x0",
		}
		assert.True(t, event.IsMint())
	})

	t.Run("isn't mint if From mismatch", func(t *testing.T) {
		event := model.Event{
			Type: model.EventTypeTransfer,
			From: config.C.Chain["ethereum"].ContractAddress,
		}
		assert.False(t, event.IsMint())
	})

	t.Run("isn't mint if Type mismatch", func(t *testing.T) {
		event := model.Event{
			Type: model.EventTypePublish,
			From: "0x0",
		}
		assert.False(t, event.IsMint())
	})
}

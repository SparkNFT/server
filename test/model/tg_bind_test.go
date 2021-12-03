package model

import (
	"testing"

	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/model"
	"github.com/stretchr/testify/assert"
)

func Test_TokenScanURL(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		tb := model.TelegramBind{
			ERC20Address: "0x123",
		}
		result, err := tb.TokenScanURL()
		assert.Nil(t, err)
		assert.Equal(t,
			config.C.Telegram.BlockViewerURLBase + "/address/" + tb.ERC20Address,
			result,
		)
	})
}

func Test_CreateTGBind(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)

		tb, err := model.CreateTGBind(int64(1337), "0x0000000000000000000000000000000000000000", "Ethereum", "ETH")
		assert.Nil(t, err)
		assert.Greater(t, tb.Id, uint64(0))
	})

	t.Run("fail if ETH address invalid", func(t *testing.T) {
		before_each(t)
		tb, err := model.CreateTGBind(int64(1337), "0xabc123", "Test", "")
		assert.Nil(t, tb)
		assert.Contains(t, err.Error(), "invalid address")
	})
}

func Test_FindTGBindBy(t *testing.T) {
	t.Run("AdminID", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			before_each(t)
			adminID := int64(1337)
			tb, _ := model.CreateTGBind(adminID, "0x0000000000000000000000000000000000000000", "", "")
			foundTBs, err := model.FindTGBindBy(&model.TelegramBind{AdminID: adminID})
			assert.Nil(t, err)
			assert.Equal(t, 1, len(foundTBs))
			foundTB := foundTBs[0]
			assert.Equal(t, tb.Id, foundTB.Id)
		})

		t.Run("not found", func(t *testing.T) {
			before_each(t)
			foundTBs, err := model.FindTGBindBy(&model.TelegramBind{AdminID: int64(1)})
			assert.Nil(t, err)
			assert.Equal(t, 0, len(foundTBs))
		})
	})
}

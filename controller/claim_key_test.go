package controller

import (
	"testing"

	"github.com/SparkNFT/key_server/config"
	"github.com/stretchr/testify/assert"
)

func before_each(t *testing.T) {
	config.ConfigPath = "../config/config.test.json"
	config.Init()
}

func Test_claim_key_check_nft(t *testing.T) {
	chainName := "ethereum"
	t.Run("success", func(t *testing.T) {
		before_each(t)
		err := claim_key_check_nft(chainName, "0x0000004215285644116B17436372D569A4ED3A1D", 4294967297)
		assert.Nil(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		before_each(t)
		err := claim_key_check_nft(chainName, "0x0000004215285644116B17436372D569A4ED3A1D", 1)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("not owned", func(t *testing.T) {
		before_each(t)
		err := claim_key_check_nft(chainName, "0x0000004215285644116B17436372D569A4ED3A10", 4294967297)
		assert.Contains(t, err.Error(), "not owned")
	})
}

func Test_claim_key_check_signature(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req := ClaimKeyRequest{
			Account: "0xbb137c332cecbc8844a009f5ede4493085f81846",
			NFTId: "21474836481",
			Signature: "0x44cdf673f4261846803dc12d9427246c49a1141573cf371f685775a7059ea0b17d6fcae84b0d371adeefb30676f00819d63e4b5bfa17775ec03f6d4ed7eb563a1b",
		}

		assert.True(t, claim_key_check_signature(&req))
	})

	t.Run("fail", func (t *testing.T) {
		req := ClaimKeyRequest{
			Account: "0xbb137c332cecbc8844a009f5ede4493085f81846",
			NFTId: "21474836480",
			Signature: "0x44cdf673f4261846803dc12d9427246c49a1141573cf371f685775a7059ea0b17d6fcae84b0d371adeefb30676f00819d63e4b5bfa17775ec03f6d4ed7eb563a1b",
		}
		assert.False(t, claim_key_check_signature(&req))
	})
}

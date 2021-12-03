package chain

import (
	"os"
	"testing"

	"github.com/SparkNFT/key_server/abi"
	"github.com/SparkNFT/key_server/chain"
	"github.com/SparkNFT/key_server/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

const (
	ROPSTEN_SUB_INF_NFT_ID string = "18446744073709551618"
)

var (
	// This address created issue_id: 1
	Issue1_creator = common.HexToAddress("0x53278E84029130805e2b659A35d8cBc35c7a6f81")
	Contract *abi.SparkLink
	Client *ethclient.Client
)

func TestMain(m *testing.M) {
	config.ConfigPath = "../../config/config.test.json"
	config.Init()
	var err error
	Contract, Client, err = chain.Init("ethereum")
	if err != nil {
		panic(err.Error())
	}

	code := m.Run()
	os.Exit(code)
}

func Test_ValidateSignature(t *testing.T) {
	t.Run("Fail", func(t *testing.T) {
		payload := `hello`

		result, err := chain.ValidateSignature(
			payload,
			"0x382496632d251a641157eb55a881e862247e7a6154e28b91081ed200f5b690d629e27bf262b6c043d37ca42759a1620e4a8bd8400a47e62cc64df3a0df3865e41c",
			Issue1_creator,
		)
		assert.False(t, result)
		assert.Nil(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		payload := `{"account":"0xbb137c332cecbc8844a009f5ede4493085f81846","root_nft_id":47244640258}`
		signature := "0x8f315f862db6ff02de4545059bac88bd38ea443268edef0a21d5eff4c3bd4eb44dd676f7c0f0c50c78d785022ad713063773542c79ac9b91290d0ca596629d061b"
		result, err := chain.ValidateSignature(
			payload,
			signature,
			common.HexToAddress("0xbb137c332cecbc8844a009f5ede4493085f81846"),
		)
		assert.Nil(t, err)
		assert.True(t, result)
	})
}

func Test_GetParentOf(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		parent, err := Contract.GetFatherByNFTId(nil, 0x200000002)
		assert.Nil(t, err)
		assert.Equal(t, uint64(0x200000001), parent)
	})

	t.Run("already root", func(t *testing.T) {
		parent, err := Contract.GetFatherByNFTId(nil, 0x100000001)
		assert.Empty(t, parent)
		assert.Contains(t, err.Error(), "SparkLink: Root NFT doesn't have father NFT.")
	})

	t.Run("not exist", func(t *testing.T) {
		parent, err := Contract.GetFatherByNFTId(nil, 0x100000002)
		assert.Empty(t, parent)
		assert.Contains(t, err.Error(), "SparkLink: Edition is not exist.")
	})
}

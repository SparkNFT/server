package pinata

import (
	"os"
	"testing"

	"github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/pinata"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	config.ConfigPath = "../../config/config.test.json"
	config.Init()

	result := m.Run()
	os.Exit(result)
}

func Test_GenerateAndRevokeAPIKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		key, err := pinata.GenerateAPIKey("ethereum", uint64(1337))
		assert.Nil(t, err)
		assert.NotNil(t, key)
		assert.IsType(t, "", key.PinataAPIKey)
		assert.IsType(t, "", key.PinataAPISecret)
		assert.Equal(t, 20, len(key.PinataAPIKey))
		assert.Equal(t, 64, len(key.PinataAPISecret))

		result := pinata.RevokeAPIKey(key.PinataAPIKey)
		assert.True(t, result)
	})
}

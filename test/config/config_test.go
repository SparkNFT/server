package config

import (
	"encoding/json"
	"testing"

	"github.com/SparkNFT/key_server/config"
	"github.com/stretchr/testify/assert"
)

func before_each(t *testing.T) {
	config.C = config.Config{}
}

func Test_Chain(t *testing.T) {
	t.Run("multi chain", func(t *testing.T) {
		before_each(t)
		configJSON := "{\"chain\": {\"ethereum\": {\"rpc_url\": \"test_ethereum\"}, \"bsc\": {\"rpc_url\": \"test_bsc\"}}}"
		err := json.Unmarshal([]byte(configJSON), &config.C)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(config.C.Chain))
		assert.Equal(t, "test_ethereum", config.C.Chain["ethereum"].RPCUrl)
		assert.Equal(t, "test_bsc", config.C.Chain["bsc"].RPCUrl)
	})

	t.Run("no chain", func(t *testing.T) {
		before_each(t)
		configJSON := "{\"chain\": {}}"
		err := json.Unmarshal([]byte(configJSON), &config.C)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(config.C.Chain))
		_, ok := config.C.Chain["ethereum"]
		assert.False(t, ok)
	})
}

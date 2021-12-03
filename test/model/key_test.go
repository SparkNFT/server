package model

import (
	"crypto/ecdsa"
	"math/rand"
	"strconv"
	"testing"

	"github.com/SparkNFT/key_server/model"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func generate_new_wallet() (public_key_hash string, private_key_hash string) {
	private_key, _ := crypto.GenerateKey()
	private_key_bytes := crypto.FromECDSA(private_key)
	private_key_hash = hexutil.Encode(private_key_bytes)[2:]

	public_key := private_key.Public()
	public_key_ecdsa, _ := public_key.(*ecdsa.PublicKey)
	public_key_bytes := crypto.FromECDSAPub(public_key_ecdsa)
	public_key_hash = hexutil.Encode(public_key_bytes)[2:]
	return public_key_hash, private_key_hash
}

func Test_CreateKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		author_address, _ := generate_new_wallet()
		issue_id := uint64(rand.Int())
		key, err := model.CreateKey(chainName, author_address, issue_id)

		assert.Nil(t, err)
		assert.NotNil(t, key)
		assert.Greater(t, key.Id, uint64(0))
	})

	t.Run("Duplicated", func(t *testing.T) {
		before_each(t)
		author_address, _ := generate_new_wallet()
		issue_id := uint64(rand.Int())
		_, err := model.CreateKey(chainName, author_address, issue_id)
		assert.Nil(t, err)

		key, err := model.CreateKey(chainName, author_address, issue_id)
		assert.Nil(t, key)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "exists")
		assert.Contains(t, err.Error(), strconv.Itoa(int(issue_id)))
	})

}

func Test_GetKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		author_address, _ := generate_new_wallet()
		issue_id := uint64(rand.Int())
		key, err := model.CreateKey(chainName, author_address, issue_id)
		assert.Nil(t, err)

		key_string, err := model.GetKey(chainName, issue_id)
		assert.Nil(t, err)
		assert.Equal(t, key.Key, key_string)
		t.Logf("Key generated: %s", key.Key)
	})
	t.Run("not found", func(t *testing.T) {
		before_each(t)
		issue_id := uint64(rand.Int())
		key_string, err := model.GetKey(chainName, issue_id)
		assert.NotNil(t, err)
		assert.Equal(t, "", key_string)
		assert.Contains(t, err.Error(), "not found")
		assert.Contains(t, err.Error(), strconv.Itoa(int(issue_id)))
	})
}

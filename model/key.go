package model

import (
	"time"

	"github.com/SparkNFT/key_server/util"
	"golang.org/x/xerrors"
)

const (
	KEY_LENGTH = 64
)

type Key struct {
	Id      uint64 `xorm:"pk autoincr"`
	Key     string `xorm:"'key' notnull"`
	Chain   string `xorm:"'chain' notnull index"`
	Owner   string `xorm:"index"`
	NFTId   uint64 `xorm:"index"`

	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
}

// CreateKey creates an encryption key
func CreateKey(chainName string, author_address string, issue_id uint64) (key *Key, err error) {
	found := Key{Chain: chainName, NFTId: issue_id, Owner: author_address}
	has, err := Engine.Get(&found)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, xerrors.Errorf("key exists: %d", issue_id)
	}

	key_string := util.RandomStringGenerator(KEY_LENGTH)
	found.Key = key_string
	_, err = Engine.Insert(&found)

	return &found, err
}

func GetKey(chainName string, nft_id uint64) (key string, err error) {
	found := Key{Chain: chainName, NFTId: nft_id}
	has, err := Engine.Get(&found)
	if err != nil {
		return "", xerrors.Errorf("error when fetching key: %w", err)
	}
	if !has {
		return "", xerrors.Errorf("NFT not found: %d", nft_id)
	}

	return found.Key, nil
}

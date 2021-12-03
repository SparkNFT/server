package model

import (
	"strconv"
	"testing"

	"github.com/SparkNFT/key_server/model"
	"github.com/stretchr/testify/assert"
)

func insert_nft_testdata(t *testing.T) {
	nfts := []model.NFT{
		{
			NFTID:     0x300000001, // 3 children
			Parent:    0x0,
			ShillCount: 3,
			MaxShillCount: 10,
			Owner: "0xA",
		},
		{
			NFTID:     0x300000002,
			Parent:    0x300000001,
			ShillCount: 0,
			MaxShillCount: 10,
			Owner: "0xB",
		},
		{
			NFTID:     0x300000003, // 3 children
			Parent:    0x300000001,
			ShillCount: 3,
			MaxShillCount: 10,
			Owner: "0xC",
		},
		{
			NFTID:     0x300000004,
			Parent:    0x300000001,
			ShillCount: 0,
			MaxShillCount: 10,
			Owner: "0xD",
		},
		{
			NFTID:     0x300000005,
			Parent:    0x300000003,
			ShillCount: 0,
			MaxShillCount: 10,
			Owner: "0xE",
		},
		{
			NFTID:     0x300000006,
			Parent:    0x300000003,
			ShillCount: 0,
			MaxShillCount: 10,
			Owner: "0xF",
		},
		{
			NFTID:     0x300000007,
			Parent:    0x300000003,
			ShillCount: 0,
			MaxShillCount: 10,
			Owner: "0x10",
		},
	}
	affected, err := model.Engine.Insert(&nfts)
	assert.Equal(t, len(nfts), int(affected))
	assert.Nil(t, err)
}

func insert_suggest_nft_data(t *testing.T) {
	nfts := []model.NFT{
		{
			NFTID: 0x300000001,
			Parent: 0x0,
			ShillCount: 10,
			MaxShillCount: 10,
			Owner: "0xA",
		},
		{
			NFTID: 0x300000002,
			Parent: 0x300000001,
			ShillCount: 10,
			MaxShillCount: 10,
			Owner: "0xB",
		},
		{
			NFTID: 0x300000003,
			Parent: 0x300000001,
			ShillCount: 5,
			MaxShillCount: 10,
			Owner: "0xC",
		},
		{
			NFTID: 0x300000004,
			Parent: 0x300000001,
			ShillCount: 5,
			MaxShillCount: 10,
			Owner: "0xA",
		},
		{
			NFTID: 0x300000005,
			Parent: 0x300000002,
			ShillCount: 5,
			MaxShillCount: 10,
			Owner: "0xA",
		},
		{
			NFTID: 0x300000006,
			Parent: 0x300000001,
			ShillCount: 10,
			MaxShillCount: 10,
			Owner: "0xE",
		},
	}
	affected, err := model.Engine.Insert(&nfts)
	assert.Equal(t, len(nfts), int(affected))
	assert.Nil(t, err)
}

func Test_ChildrenCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_nft_testdata(t)

		// Root NFT
		result, err := model.ChildrenCount(chainName, uint64(0x300000001))
		assert.Nil(t, err)
		assert.Equal(t, 6, result)

		// NFT with less children
		result, err = model.ChildrenCount(chainName, uint64(0x300000003))
		assert.Nil(t, err)
		assert.Equal(t, 3, result)

		// NFT found, but no children
		result, err = model.ChildrenCount(chainName, uint64(0x30000002))
		assert.Nil(t, err)
		assert.Equal(t, 0, result)

		// No NFT found
		result, err = model.ChildrenCount(chainName, uint64(0x30000010))
		assert.Nil(t, err)
		assert.Equal(t, 0, result)
	})
}

func Test_ChildrenTree(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_nft_testdata(t)

		tree, err := model.ChildrenTree(chainName, uint64(0x300000001))
		assert.Nil(t, err)
		assert.Equal(t, strconv.FormatUint(uint64(0x300000001), 10), tree.NFTID)
		assert.Equal(t, 3, len(tree.Children))
		found := false
		for _, n := range tree.Children {
			if (n.NFTID == strconv.FormatUint(uint64(0x300000003), 10)) {
				assert.Equal(t, 3, len(n.Children))
				found = true
			}
		}
		assert.True(t, found)
	})
}


func Test_FindNFT(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_nft_testdata(t)
		nft_id := uint64(0x300000003)

		nft, err := model.FindNFT(chainName, nft_id)
		assert.Nil(t, err)
		assert.NotNil(t, nft)
		assert.Equal(t, nft_id, nft.NFTID)
	})
}

func Test_CanShill(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_nft_testdata(t)
		nft_id := uint64(0x300000003)
		nft, _ := model.FindNFT(chainName, nft_id)
		assert.True(t, nft.CanShill())

		nft.ShillCount = nft.MaxShillCount
		model.Engine.Update(nft)
		assert.False(t, nft.CanShill())
	})
}

func Test_ChildrenCount_InstanceMethod(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_nft_testdata(t)
		nft_id := uint64(0x300000003)
		nft, _ := model.FindNFT(chainName, nft_id)
		assert.Equal(t, int64(3), nft.ChildrenCount())

		nft_id = uint64(0x300000001)
		nft, _ = model.FindNFT(chainName, nft_id)
		assert.Equal(t, int64(3), nft.ChildrenCount())

		nft_id = uint64(0x300000002)
		nft, _ = model.FindNFT(chainName, nft_id)
		assert.Equal(t, int64(0), nft.ChildrenCount())
	})
}

func Test_Children(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_nft_testdata(t)
		nft_id := uint64(0x300000003)
		nft, _ := model.FindNFT(chainName, nft_id)
		children, err := nft.Children()
		assert.Nil(t, err)
		assert.Equal(t, 3, len(children))
	})

	t.Run("Empty result", func(t *testing.T) {
		before_each(t)
		insert_nft_testdata(t)

		nft_id := uint64(0x300000002)
		nft, _ := model.FindNFT(chainName, nft_id)
		no_children, err := nft.Children()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(no_children))
	})
}

func Test_Suggest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		insert_suggest_nft_data(t)

		nft, _ := model.FindNFT(chainName, uint64(0x300000001))
		next, err := nft.Suggest(nil)
		assert.Nil(t, err)
		assert.Equal(t, uint64(0x300000004), next.NFTID)
	})

	t.Run("returns self", func(t *testing.T) {
		before_each(t)
		insert_suggest_nft_data(t)

		nft, _ := model.FindNFT(chainName, uint64(0x300000004))
		next, err := nft.Suggest(nil)
		assert.Nil(t, err)
		assert.Equal(t, uint64(0x300000004), next.NFTID)
	})

	t.Run("returns different owner", func(t *testing.T) {
		before_each(t)
		insert_suggest_nft_data(t)

		nft, _ := model.FindNFT(chainName, uint64(0x300000002))
		next, err := nft.Suggest(nil)
		assert.Nil(t, err)
		assert.NotNil(t, next)
		assert.Equal(t, uint64(0x300000005), next.NFTID)
	})

	t.Run("returns nil", func(t *testing.T) {
		before_each(t)
		insert_suggest_nft_data(t)

		nft, _ := model.FindNFT(chainName, uint64(0x300000006))
		next, err := nft.Suggest(nil)
		assert.Nil(t, err)
		assert.Nil(t, next)
	})
}

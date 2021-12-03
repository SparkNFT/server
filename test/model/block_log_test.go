package model

import (
	"math/rand"
	"testing"

	"github.com/SparkNFT/key_server/model"
	"github.com/stretchr/testify/assert"
)

func Test_BlockLogStart(t *testing.T) {
	t.Run("duplicated", func (t *testing.T) {
		before_each(t)
		session := model.Engine.NewSession()
		_ = session.Begin()
		defer session.Close()

		height := uint64(rand.Uint32())

		err := model.BlockLogStart(session, chainName, height)
		assert.Nil(t, err)

		err = model.BlockLogStart(session, chainName, height)
		session.Commit()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")
	})

	t.Run("success", func(t *testing.T) {
		before_each(t)
		session := model.Engine.NewSession()
		_ = session.Begin()
		defer session.Close()
		height := uint64(rand.Uint32())
		err := model.BlockLogStart(session, chainName, height)
		assert.Nil(t, err)
		session.Commit()

		block := model.BlockLog{BlockHeight: height}
		found, err := model.Engine.Get(&block)
		assert.Nil(t, err)

		assert.True(t, found)
		assert.Greater(t, block.Id, uint64(0))
		assert.False(t, block.Scanned)
	})
}

func Test_BlockLogFinish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		session := model.Engine.NewSession()
		_ = session.Begin()
		defer session.Close()

		height := uint64(rand.Uint32())
		err := model.BlockLogStart(session, chainName, height)
		assert.Nil(t, err)

		err = model.BlockLogFinish(session, chainName, height)
		assert.Nil(t, err)
		session.Commit()

		block := model.BlockLog{BlockHeight: height}
		found, err := model.Engine.Get(&block)
		assert.Nil(t, err)
		assert.True(t, found)
		assert.Greater(t, block.Id, uint64(0))
		assert.True(t, block.Scanned)
	})

	t.Run("not_found", func(t *testing.T) {
		before_each(t)
		session := model.Engine.NewSession()
		_ = session.Begin()
		defer session.Close()

		height := uint64(rand.Uint32())
		err := model.BlockLogFinish(session, chainName, height)
		session.Commit()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func Test_BlockLogFindFirst(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		before_each(t)
		session := model.Engine.NewSession()
		_ = session.Begin()
		defer session.Close()

		height := uint64(rand.Uint32())
		model.BlockLogStart(session, chainName, height)
		model.BlockLogFinish(session, chainName, height)
		session.Commit()

		found, err := model.BlockLogFindFirst(chainName)
		assert.Nil(t, err)
		assert.Equal(t, height, found.BlockHeight)
	})

	t.Run("omit unfinished", func(t *testing.T) {
		before_each(t)
		session := model.Engine.NewSession()
		_ = session.Begin()
		defer session.Close()

		height := uint64(rand.Uint32())
		model.BlockLogStart(session, chainName, height)
		model.BlockLogStart(session, chainName, height + 1)
		model.BlockLogFinish(session, chainName, height)
		session.Commit()

		found, err := model.BlockLogFindFirst(chainName)
		assert.Nil(t, err)
		assert.Equal(t, height, found.BlockHeight)
	})

	t.Run("not found", func(t *testing.T) {
		before_each(t)
		session := model.Engine.NewSession()
		_ = session.Begin()
		defer session.Close()

		height := uint64(rand.Uint32())
		model.BlockLogStart(session, chainName, height)
		session.Commit()

		found, err := model.BlockLogFindFirst(chainName)
		assert.NotNil(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "found height failed")
	})
}

func Test_BlockLogClean(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		log := model.BlockLog{
			BlockHeight: 1000,
			Scanned:     true,
		}
		fetched := model.BlockLog{}
		affected, err := model.Engine.Insert(&log)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), affected)

		err = model.BlockLogClean(chainName)
		assert.Nil(t, err)

		model.Engine.ID(log.Id).Get(&fetched)
		assert.Equal(t, log.Id, fetched.Id)
	})

	t.Run("success with log cleaned", func(t *testing.T) {
		before_each(t)

		for i := uint64(0); i < 150; i++ {
			model.Engine.Insert(model.BlockLog{
				BlockHeight: i + 1000,
				Scanned: true,
			})
		}

		err := model.BlockLogClean(chainName)
		assert.Nil(t, err)

		count, err := model.Engine.Count(&model.BlockLog{})
		assert.Nil(t, err)
		assert.Equal(t, int64(101), count)
	})
}

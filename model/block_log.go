package model

import (
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type BlockLog struct {
	Id          uint64    `xorm:"pk autoincr"`
	Chain       string    `xorm:"'chain' index notnull"`
	BlockHeight uint64    `xorm:"'block_height' notnull"`
	Scanned     bool      `xorm:"'scanned' index default(false)"`

	CreatedAt   time.Time `xorm:"'created_at' created"`
	UpdatedAt   time.Time `xorm:"'updated_at' updated"`
}

func (BlockLog) TableName() string {
	return "block_logs"
}

func BlockLogStart(session *xorm.Session, chainName string, height uint64) (err error) {
	bl := &BlockLog{
		Chain: chainName,
		BlockHeight: height,
		Scanned:     false,
	}
	affected, err := session.Insert(bl)
	if err != nil {
		return xerrors.Errorf("error when starting a BlockLog at height %d: %w", height, err)
	}
	if affected == 0 {
		return xerrors.Errorf("Error during inserting BlockLog: nothing inserted")
	}
	logrus.WithFields(logrus.Fields{"chain": chainName, "height": height}).WithField("model", "BlockLog").Debugf("Height written in DB")

	return nil
}

func BlockLogFinish(session *xorm.Session, chainName string, height uint64) (err error) {
	affected, err := session.Cols("scanned").Update(BlockLog{
		Scanned: true,
	}, BlockLog{
		Chain: chainName,
		BlockHeight: height,
	})
	if err != nil {
		return xerrors.Errorf("error when updating block log %d: %w", height, err)
	}
	if affected == 0 {
		return xerrors.Errorf("Block height %d not found", height)
	}
	logrus.WithFields(logrus.Fields{"chain": chainName, "height": height}).WithField("model", "BlockLog").Debugf("Height finished in DB")

	return nil
}

func BlockLogFindFirst(chainName string) (result *BlockLog, err error) {
	result = &BlockLog{}
	found, err := Engine.Where(builder.Eq{"scanned": true, "chain": chainName}).Desc("block_height").Get(result)
	if err != nil {
		return nil, xerrors.Errorf("error when finding first block log: %w", err)
	}
	if !found {
		return nil, xerrors.Errorf("found height failed")
	}
	logrus.WithFields(logrus.Fields{"chain": chainName, "height": result.BlockHeight}).WithField("model", "BlockLog").Debugf("Height found in DB.")

	return result, nil
}

func BlockLogClean(chainName string) (err error) {
	latest_log, err := BlockLogFindFirst(chainName)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	id := latest_log.Id
	if id <= uint64(100) {
		return nil
	}

	// TODO: reserved log amount should be configurable
	_, err = Engine.Where(builder.Lt{"id": (id - 100)}).And(builder.Eq{"chain": chainName}).Delete(&BlockLog{})
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

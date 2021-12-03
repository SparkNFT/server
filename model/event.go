package model

import (
	"time"

	"github.com/SparkNFT/key_server/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"xorm.io/xorm"
)

type EventType string

const (
	EventTypeTransfer EventType = "Transfer"
	EventTypePublish  EventType = "Publish"
)

type Event struct {
	Id          uint64 `xorm:"pk autoincr"`
	Chain       string `xorm:"'chain' index notnull"`
	BlockHeight uint64 `xorm:"'block_height' index notnull"`
	Index       uint   `xorm:"'event_index'"`
	TxHash      string `xorm:"'tx_hash'"`
	TxIndex     uint   `xorm:"'tx_index'"`

	Type      EventType `xorm:"'type' index notnull"`
	From      string    `xorm:"'from' index notnull"`
	To        string    `xorm:"'to' index notnull"`
	NFTId     uint64    `xorm:"'nft_id' index notnull"`
	TokenAddr string    `xorm:"'token_addr' index"`

	CreatedAt time.Time `xorm:"'created_at' created"`
	UpdatedAt time.Time `xorm:"'updated_at' updated"`
}

func (Event) TableName() string {
	return "events"
}

func (event Event) IsPublish() bool {
	return event.Type == EventTypePublish
}

func (event Event) IsTransfer() bool {
	return event.Type == EventTypeTransfer
}

func (event Event) FromAddress() common.Address {
	return common.HexToAddress(event.From)
}

func (event Event) ToAddress() common.Address {
	return common.HexToAddress(event.To)
}

func (event Event) IsRoot() bool {
	return (event.IsPublish()) || (event.FromAddress() == common.HexToAddress("0x0"))
}

func (event Event) IsMint() bool {
	return event.IsTransfer() && (event.FromAddress() == common.HexToAddress("0x0"))
}

// NFTId = (IssueId(32bit) << 32) | EditionId(32bit)
func (event Event) IssueId() uint32 {
	return uint32(event.NFTId >> 32)
}

// NFTId = (IssueId(32bit) << 32) | EditionId(32bit)
func (event Event) EditionId() uint32 {
	return uint32(event.NFTId)
}

func CreateFromBlockEventPublish(session *xorm.Session, chainName string, logs []abi.SparkLinkPublish) (events []*Event, err error) {
	l := logrus.WithFields(logrus.Fields{"chain": chainName, "count": len(logs), "model": "Event", "type": "Publish"})
	if len(logs) == 0 {
		l.Debugf("No Publish event parsed.")
		return nil, nil
	}

	l.Debugf("Start parsing logs")
	events = make([]*Event, 0, len(logs))
	for _, log := range logs {
		event := &Event{
			Chain:       chainName,
			BlockHeight: log.Raw.BlockNumber,
			Index:       log.Raw.Index,
			TxHash:      log.Raw.TxHash.Hex(),
			TxIndex:     log.Raw.TxIndex,
			Type:        EventTypePublish,
			From:        common.HexToAddress("0x0").Hex(),
			To:          log.Publisher.Hex(),
			NFTId:       log.RootNFTId,
			TokenAddr:   log.TokenAddr.Hex(),
		}
		events = append(events, event)
	}
	l.Debugf("Publish event ready to be inserted: %+v", events)

	affected, err := session.Insert(events)
	l.WithField("affected", affected).Debugf("Insert finished")
	if err != nil || int(affected) != len(events) {
		return nil, xerrors.Errorf("error when inserting Publish Event to DB: %w", err)
	}

	return events, nil
}

func CreateFromBlockEventTransfer(session *xorm.Session, chainName string, logs []abi.SparkLinkTransfer) (events []*Event, err error) {
	l := logrus.WithFields(logrus.Fields{"chain": chainName, "count": len(logs), "model": "Event", "type": "Transfer"})
	if len(logs) == 0 {
		l.Debugf("No Transfer event parsed.")
		return nil, nil
	}

	events = make([]*Event, 0, len(logs))
	l.Debugf("Start parsing logs")
	for _, log := range logs {
		event := &Event{
			Chain:       chainName,
			BlockHeight: log.Raw.BlockNumber,
			Index:       log.Raw.Index,
			TxHash:      log.Raw.TxHash.Hex(),
			TxIndex:     log.Raw.TxIndex,
			Type:        EventTypeTransfer,
			From:        log.From.Hex(),
			To:          log.To.Hex(),
			NFTId:       log.TokenId.Uint64(),
			TokenAddr:   "",
		}
		events = append(events, event)
	}

	affected, err := session.Insert(events)
	l.WithField("affected", affected).Debugf("Insert finished")
	if err != nil || int(affected) != len(logs) {
		return nil, xerrors.Errorf("error when inserting Transfer Event to DB: %w", err)
	}

	return events, nil
}

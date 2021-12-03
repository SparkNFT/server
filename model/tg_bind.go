package model

import (
	"net/url"
	"time"

	"github.com/SparkNFT/key_server/config"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/xerrors"
	"xorm.io/builder"
)

type TelegramBind struct {
	Id           uint64 `xorm:"'id' BIGINT pk autoincr"`
	ERC20Name    string `xorm:"'erc20_name' notnull"`
	ERC20Symbol  string `xorm:"'erc20_symbol' notnull"`
	ERC20Address string `xorm:"'erc20_address' notnull"`
	AdminID      int64  `xorm:"'admin_id' index notnull"`

	CreatedAt time.Time `xorm:"'created_at' created"`
	UpdatedAt time.Time `xorm:"'updated_at' updated"`
}

func (TelegramBind) TableName() string {
	return "telegram_bind"
}

func (tb *TelegramBind) TokenScanURL() (result string, err error) {
	url, err := url.Parse(config.C.Telegram.BlockViewerURLBase + "/address/" + tb.ERC20Address)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return url.String(), nil
}

func (tb *TelegramBind) TelegramGroups() (result []*TelegramGroup, err error) {
	result = make([]*TelegramGroup, 0)
	err = Engine.Where(builder.Eq{"tg_bind_id": tb.Id}).Find(&result)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	return result, nil
}

func CreateTGBind(adminID int64, erc20addr, erc20name, erc20symbol string) (instance *TelegramBind, err error) {
	instance = &TelegramBind{
		AdminID:      adminID,
		ERC20Address: erc20addr,
		ERC20Name:    erc20name,
		ERC20Symbol:  erc20symbol,
	}
	if !common.IsHexAddress(erc20addr) {
		return nil, xerrors.Errorf("invalid address: %s", erc20addr)
	}
	count, err := Engine.Insert(instance)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	if count != int64(1) {
		return nil, xerrors.Errorf("insert error: count is %d", count)
	}

	return instance, nil
}

// FindTGBindBy returns a slice of results for given condition struct
func FindTGBindBy(condition *TelegramBind) (results []*TelegramBind, err error) {
	results = make([]*TelegramBind, 0)
	err = Engine.Find(&results, condition)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func DeleteTGBind(binds []*TelegramBind) (affected int64, err error) {
	ids := make([]uint64, 0)
	for _, tb := range binds {
		ids = append(ids, tb.Id)
		DeleteTGGroupByBind(tb)
	}

	return Engine.Where(builder.In("id", ids)).Delete(&TelegramBind{})
}

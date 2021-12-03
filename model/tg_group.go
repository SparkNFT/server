package model

import (
	"golang.org/x/xerrors"
	"xorm.io/builder"
)

type TelegramGroup struct {
	Id           uint64 `xorm:"'id' BIGINT pk autoincr"`
	TGBindID     uint64 `xorm:"'tg_bind_id' BIGINT index"`
	ChatID       int64  `xorm:"'chat_id'"`
	ChatTitle    string `xorm:"'chat_title'"`
}

func (TelegramGroup) TableName() string {
	return "telegram_group"
}

func (tg *TelegramGroup) TelegramBind() (result *TelegramBind, err error) {
	result = &TelegramBind{}
	found, err := Engine.Where(builder.Eq{"id": tg.Id}).Get(result)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	if !found {
		return nil, nil
	}
	return result, nil
}

func CreateTGGroup(bind *TelegramBind, chatId int64, chatTitle string) (group *TelegramGroup, err error) {
	group = &TelegramGroup{
		TGBindID: bind.Id,
		ChatID: chatId,
		ChatTitle: chatTitle,
	}

	affected, err := Engine.Insert(group)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	if affected != 1 {
		return nil, xerrors.Errorf("TelegramGroup insert failed")
	}
	return group, nil
}

func DeleteTGGroupByBind(bind *TelegramBind) (affected int64, err error) {
	return Engine.Where(builder.Eq{"tg_bind_id": bind.Id}).Delete(&TelegramGroup{})
}

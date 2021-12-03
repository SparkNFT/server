package telegram

import (
	"fmt"
	"strings"

	"github.com/SparkNFT/key_server/chain"
	"github.com/SparkNFT/key_server/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/looplab/fsm"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	tele "gopkg.in/tucnak/telebot.v3"
)

type ConversationStatus string
type ConversationEvent string

var (
	Conversations = make(map[int64]*Conversation, 0)
)

const (
	STop                ConversationStatus = "top"
	SBindStarted        ConversationStatus = "bind_started"
	SBindInputedERC20   ConversationStatus = "bind_inputed_erc20"
	SUnbindStarted      ConversationStatus = "unbind_started"
	SUnbindInputedERC20 ConversationStatus = "unbind_inputed_erc20"

	EBackToTop          ConversationEvent = "back_to_top"
	EBindStart          ConversationEvent = "bind_start"
	EBindInputERC20     ConversationEvent = "bind_input_erc20"
	EBindConfirmERC20   ConversationEvent = "bind_confirm_erc20"
	EListERC20Binding   ConversationEvent = "list_erc20_binding"
	EUnbindStart        ConversationEvent = "unbind_start"
	EUnbindInputERC20   ConversationEvent = "unbind_input_erc20"
	EUnbindConfirmERC20 ConversationEvent = "unbind_confirm_erc20"
)

type Conversation struct {
	User         int64
	Erc20Address string
	Erc20Name    string
	Erc20Symbol  string
	FSM          *fsm.FSM
	Error        error
	Tele         tele.Context
}

type ArgEnterInputedERC20 struct {
	Erc20Address string
}

func GetUsernameInContext(m tele.Context) int64 {
	return m.Message().Sender.ID
}

func GetConversation(m tele.Context) (stateMachine *Conversation) {
	user := GetUsernameInContext(m)
	conv, ok := Conversations[user]
	if ok && conv != nil {
		conv.Tele = m
		return conv
	}

	conv = &Conversation{
		User:  user,
		Error: nil,
		Tele:  m,
	}
	conv.FSM = fsm.NewFSM(
		string(STop),
		fsm.Events{
			{
				Name: string(EBackToTop),
				Src:  []string{string(STop), string(SBindStarted), string(SBindInputedERC20), string(SUnbindInputedERC20), string(SUnbindStarted)},
				Dst:  string(STop),
			},
			{
				Name: string(EBindStart),
				Src:  []string{string(STop), string(SBindInputedERC20)},
				Dst:  string(SBindStarted),
			},
			{
				Name: string(EBindInputERC20),
				Src:  []string{string(SBindStarted)},
				Dst:  string(SBindInputedERC20),
			},
			{
				Name: string(EBindConfirmERC20),
				Src:  []string{string(SBindInputedERC20)},
				Dst:  string(STop),
			},
			{
				Name: string(EListERC20Binding),
				Src:  []string{string(STop)},
				Dst:  string(STop),
			},
			{
				Name: string(EUnbindStart),
				Src:  []string{string(STop)},
				Dst:  string(SUnbindStarted),
			},
			{
				Name: string(EUnbindInputERC20),
				Src:  []string{string(SUnbindStarted)},
				Dst:  string(SUnbindInputedERC20),
			},
			{
				Name: string(EUnbindConfirmERC20),
				Src:  []string{string(SUnbindInputedERC20)},
				Dst:  string(STop),
			},
		},
		fsm.Callbacks{
			enter(STop):                 conv.top,
			before(EBindInputERC20):     conv.before_bind_input_erc20,
			before(EBindConfirmERC20):   conv.before_bind_confirm_erc20,
			before(EListERC20Binding):   conv.list_erc20_binding,
			before(EUnbindInputERC20):   conv.before_unbind_input_erc20,
			before(EUnbindConfirmERC20): conv.before_unbind_confirm_erc20,
		},
	)
	Conversations[user] = conv
	return conv
}

// EmitEvent triggers a status change event
func (conv *Conversation) EmitEvent(event ConversationEvent) (err error) {
	conv.Error = conv.FSM.Event(string(event))
	return conv.Error
}

// Text returns bot chat message corresponding to current conversation status.
func (conv *Conversation) Reply() (err error) {
	if conv.Error != nil {
		conv.Tele.EditOrReply(
			fmt.Sprintf("Error occured: %s.\nBack to start.", conv.Error.Error()),
			Menus["main"],
		)
		// Ignore BackToTop event error anyway
		conv.FSM.Event(string(EBackToTop))
		return nil
	}

	switch conv.FSM.Current() {
	case string(STop):
		conv.Tele.EditOrReply("hello", Menus["main"])
	case string(SBindStarted):
		conv.Tele.EditOrReply("OK. Give me your ERC20 token contract address (0xAbCD....):", Menus["back"])
	case string(SBindInputedERC20):
		conv.Tele.EditOrReply(
			fmt.Sprintf("You're now adding %s ($%s) with contract address %s.\nAre you sure?",
				conv.Erc20Name,
				conv.Erc20Symbol,
				conv.Erc20Address,
			),
			Menus["confirm"],
		)
	case string(SUnbindStarted):
		conv.Tele.EditOrReply("OK. Give me your ERC20 token contract address (0xAbCD....):", Menus["back"])
	case string(SUnbindInputedERC20):
		conv.Tele.EditOrReply(
			fmt.Sprintf("You're now unlinking contract address\n%s\n\nAre you sure?", conv.Erc20Address),
			Menus["confirm"],
		)
	default:
		conv.Tele.EditOrReply("Unrecognized conversation status.\nBack to start.", Menus["main"])
	}
	return nil
}

func (conv *Conversation) SelfDestroy() {
	delete(Conversations, conv.User)
}

func (conv *Conversation) top(_ *fsm.Event) {
	conv.Erc20Address = ""
	conv.Error = nil
	log.WithFields(logrus.Fields{"user": conv.User}).Debugf("Top")
}

func (conv *Conversation) before_bind_input_erc20(event *fsm.Event) {
	if !common.IsHexAddress(conv.Tele.Text()) {
		conv.Error = xerrors.Errorf("invalid ERC20 address")
		event.Cancel(conv.Error)
		return
	}
	conv.Erc20Address = conv.Tele.Text()
	conv.Tele.Reply(fmt.Sprintf("Now querying ERC20 info of %s", conv.Erc20Address))
	name, symbol, err := chain.GetERC20Info(conv.Erc20Address)
	if err != nil {
		conv.Error = xerrors.Errorf("get ERC20 info failed: %w", err)
		event.Cancel(conv.Error)
		return
	}
	conv.Erc20Name = name
	conv.Erc20Symbol = symbol
	conv.LogDebug("Inputed ERC20")
}

func (conv *Conversation) before_bind_confirm_erc20(event *fsm.Event) {
	_, err := model.CreateTGBind(conv.User, conv.Erc20Address, conv.Erc20Name, conv.Erc20Symbol)
	if err != nil {
		conv.Error = err
		event.Cancel(conv.Error)
	}
	conv.LogDebug("Saved ERC20 to DB. Error: %+v", err)
}

func (conv *Conversation) list_erc20_binding(_ *fsm.Event) {
	tg_binds, err := model.FindTGBindBy(&model.TelegramBind{AdminID: conv.AdminID()})
	if err != nil {
		conv.Error = err
		conv.Reply()
	}
	if len(tg_binds) == 0 {
		conv.Tele.EditOrReply("You have no ERC20 bindings.", Menus["main"])
		return
	}

	result := make([]string, 0)
	result = append(result, "Currently binded ERC20 by you:")
	for _, tg_bind := range tg_binds {
		token_scan_url, _ := tg_bind.TokenScanURL()
		result = append(result, fmt.Sprintf("[%s](%s)", tg_bind.ERC20Address, token_scan_url))
	}
	conv.Tele.EditOrReply(strings.Join(result[:], "\n"), Menus["main"], &tele.SendOptions{
		DisableWebPagePreview: true,
		DisableNotification:   true,
		ParseMode:             "markdownv2",
	})
}

func (conv *Conversation) before_unbind_input_erc20(event *fsm.Event) {
	erc20 := conv.Tele.Text()
	if !common.IsHexAddress(erc20) {
		conv.Error = xerrors.Errorf("invalid ERC20 address")
		event.Cancel(conv.Error)
		return
	}

	found, err := model.Engine.Exist(&model.TelegramBind{ERC20Address: erc20, AdminID: conv.AdminID()})
	if err != nil {
		conv.Error = xerrors.Errorf("error when fetching ERC20 record: %w", err)
		event.Cancel(conv.Error)
		return
	}
	if !found {
		conv.Error = xerrors.Errorf("this address is not found in your ERC20 binding list.")
		event.Cancel(conv.Error)
		return
	}

	conv.Erc20Address = erc20
}

func (conv *Conversation) before_unbind_confirm_erc20(event *fsm.Event) {
	results, err := model.FindTGBindBy(&model.TelegramBind{
		ERC20Address: conv.Erc20Address,
		AdminID:      conv.AdminID(),
	})
	if err != nil {
		conv.Error = xerrors.Errorf("error when fetching ERC20 record: %w", err)
		event.Cancel(conv.Error)
		return
	}

	count, err := model.DeleteTGBind(results)
	if err != nil {
		conv.Error = xerrors.Errorf("error when deleting binding: %w", err)
		event.Cancel(conv.Error)
		return
	}
	conv.LogDebug("Deleted ERC20 binding. count: %d", count)
}

func (conv *Conversation) LogDebug(format string, args ...interface{}) {
	log.WithFields(logrus.Fields{
		"ERC20": conv.Erc20Address,
		"User":  conv.User,
	}).Debugf(format, args...)
}

func (conv *Conversation) AdminID() (adminID int64) {
	return GetUsernameInContext(conv.Tele)
}

// helper functions
func before(event ConversationEvent) string {
	return fmt.Sprintf("before_%s", event)
}

func enter(state ConversationStatus) string {
	return fmt.Sprintf("enter_%s", state)
}

func after(event ConversationEvent) string {
	return fmt.Sprintf("after_%s", event)
}

package telegram

import (
	"fmt"

	tele "gopkg.in/tucnak/telebot.v3"
)

var (
	btnBindERC20 = tele.InlineButton{Text: "🔗 Link ERC20", Unique: "link_erc20"}
	btnListERC20 = tele.InlineButton{Text: "📃 List binded", Unique: "list_binded_erc20"}
	btnUnbindERC20 = tele.InlineButton{Text: "🛇 Unlink ERC20", Unique: "unlink_erc20"}

	btnPing = tele.InlineButton{Text: "ℹ Ping", Unique: "ping"}
	btnTop = tele.InlineButton{Text: "🏠 Back to top", Unique: "top"}

	btnYes = tele.InlineButton{Text: "✓ Yes", Unique: "yes"}
)

func initMenus() {
	// Main menu
	main := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		InlineKeyboard: [][]tele.InlineButton{
			{btnBindERC20},
			{btnListERC20},
			{btnUnbindERC20},
		},
	}
	Menus["main"] = main
	B.Handle(&btnPing, ping)
	B.Handle(&btnBindERC20, bindERC20)
	B.Handle(&btnListERC20, listERC20)
	B.Handle(&btnUnbindERC20, unbindERC20)

	// Confirm menu
	confirm := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		InlineKeyboard: [][]tele.InlineButton{
			{ btnYes, btnTop },
		},
	}
	Menus["confirm"] = confirm
	B.Handle(&btnYes, handleYes)

	// Back to start menu
	back := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		InlineKeyboard: [][]tele.InlineButton{
			{ btnTop },
		},
	}
	Menus["back"] = back
	B.Handle(&btnTop, start)
}

func renderMenuOnText(m tele.Context) (err error) {
	conv := GetConversation(m)
	// the state BEFORE user-given action
	switch ConversationStatus(conv.FSM.Current()) {
	case SBindStarted:
		conv.EmitEvent(EBindInputERC20)
		conv.Reply()
	case SUnbindStarted:
		conv.EmitEvent(EUnbindInputERC20)
		conv.Reply()
	default:
		log.Warnf("receiving message in status %+v", conv.FSM.Current())
		m.Send(fmt.Sprintf("Unknown input: %s\nBack to top.", m.Text()), Menus["main"])
		conv.EmitEvent(EBackToTop)
	}
	return nil
}

func handleYes(m tele.Context) (err error) {
	conv := GetConversation(m)
	switch ConversationStatus(conv.FSM.Current()) {
	case SBindInputedERC20:
		conv.EmitEvent(EBindConfirmERC20)
	case SUnbindInputedERC20:
		conv.EmitEvent(EUnbindConfirmERC20)
	default:
		m.EditOrReply("Unknown conversation status.\nBack to top.", Menus["main"])
		conv.EmitEvent(EBackToTop)
	}
	return nil
}

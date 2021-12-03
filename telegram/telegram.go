package telegram

import (
	"fmt"
	"time"

	"github.com/SparkNFT/key_server/config"
	l "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	tele "gopkg.in/tucnak/telebot.v3"
)

var (
	B     *tele.Bot
	log   *l.Entry
	Menus = make(map[string]*tele.ReplyMarkup, 0)
)

func Init(offline bool) {
	if log == nil {
		log = l.WithFields(l.Fields{
			"module": "telegram",
		})
	}

	if B == nil {
		initBot(offline)
		log.WithFields(l.Fields{
			"username": B.Me.Username,
			"IsBot":    B.Me.IsBot,
		}).Infof("Telegram Bot initialized.")
	}
	initMenus()

	B.Handle("/start", start)
	B.Handle(tele.OnAddedToGroup, addedToGroup)

	// TODO: move these into separate menus
	B.Handle("/issue", showIssue)
	B.Handle("/issues", listIssue)

	// Fallback
	B.Handle(tele.OnText, onText)
}

func initBot(offline bool) {
	var err error
	B, err = tele.NewBot(tele.Settings{
		Token: config.C.Telegram.Token,
		Poller: &tele.LongPoller{
			Timeout: 10 * time.Second,
		},
		Offline: offline,
	})
	if err != nil {
		panic(err)
	}
}

func parseError(m tele.Context, err error) error {
	switch err {
	case tele.ErrBotKickedFromGroup, tele.ErrBotKickedFromSuperGroup, tele.ErrKickingChatOwner:
		log.WithFields(l.Fields{
			"chat_id": m.Chat().Recipient(),
		}).Warnf("Bot was kicked from group: %+v", err)
		// TODO: Do sth in Model side
		return nil
	default:
		return err
	}
}

func ping(c tele.Context) (err error) {
	return c.EditOrSend(fmt.Sprintf("Hello %s", c.Chat().Username))
}

func test(c tele.Context) (err error) {
	log.Debugf("Test Data: %+v", c.Data())
	return c.EditOrReply("Test page", &tele.SendOptions{
		ReplyMarkup:  Menus["test"],
	})
}

func test2(c tele.Context) (err error) {
	log.Debugf("Test2 Data: %+v", c.Data())
	return c.EditOrReply("Test page 2", &tele.SendOptions{
		ReplyMarkup:  Menus["test2"],
	})
}


func listIssue(m tele.Context) (err error) {
	err = m.Reply("TODO")
	return parseError(m, err)
}

func showIssue(m tele.Context) (err error) {
	err = m.Reply("TODO")
	return parseError(m, err)
}

func bindERC20(m tele.Context) (err error) {
	conv := GetConversation(m)
	conv.EmitEvent(EBindStart)
	return conv.Reply()
}

func unbindERC20(m tele.Context) (err error) {
	conv := GetConversation(m)
	conv.EmitEvent(EUnbindStart)
	return conv.Reply()
}

func listERC20(m tele.Context) (err error) {
	conv := GetConversation(m)
	conv.EmitEvent(EListERC20Binding)
	return nil
}

func addedToGroup(m tele.Context) (err error) {
	sendGreetingMessage(m)
	return nil
}

func start(m tele.Context) (err error) {
	if !m.Message().Private() {
		m.Reply("Please call me in private chat")
		return xerrors.Errorf("telegram: /start should be called in private chat")
	}
	conv := GetConversation(m)
	conv.EmitEvent(EBackToTop)
	conv.Reply()

	return nil
}

func onText(m tele.Context) (err error) {
	if !m.Message().Private() {
		return nil
	}
	return renderMenuOnText(m)
}

func onInline(m tele.Context) (err error) {
	log.Debugf("Inline: context: %+v", m)
	return nil
}

func onQuery(m tele.Context) (err error) {
	log.Debugf("Query: context: %+v", m)
	return nil
}

func getChatTitle(chatID int64) (name string, err error) {
	chat, err := B.ChatByID(chatID)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return chat.Title, nil
	// B.Send(to tele.Recipient, what interface{}, opts ...interface{})
}

// contextIsAdmin detects if the message sender in current chat is creator or admin role.
func contextIsAdmin(m tele.Context) (result bool, err error) {
	member, err := B.ChatMemberOf(m.Chat(), m.Sender())
	if err != nil {
		return false, xerrors.Errorf("%w", err)
	}

	return (member.Role == tele.Administrator || member.Role == tele.Creator), nil
}

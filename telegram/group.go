package telegram

import (
	"fmt"
	"strings"

	"github.com/SparkNFT/key_server/model"
	"golang.org/x/xerrors"
	tele "gopkg.in/tucnak/telebot.v3"
)

func sendGreetingMessage(m tele.Context) {
	is_admin, err := contextIsAdmin(m)
	if err != nil {
		sendErrorMessage(m, "Error when getting current user role", err)
		return
	}
	if !is_admin {
		sendErrorMessage(m, "Only Creator or Administrator of this group can invite me", xerrors.Errorf("wrong role when inviting bot in group"))
	}

	binds, err := model.FindTGBindBy(&model.TelegramBind{AdminID: m.Message().Sender.ID})
	if err != nil {
		sendErrorMessage(m, "Error when finding existed ERC20 binding", err)
		return
	}
	if len(binds) == 0 {
		sendErrorMessage(
			m,
			"Seems there's no ERC20 binded by my invitor. For admin of this group, please DM me first to bind at least one ERC20 address",
			xerrors.Errorf("no ERC20 binded"),
		)
		return
	}

	message := strings.Builder{}
	defer message.Reset()
	message.WriteString("Hello everyone. I'm <b>SparkLink</b> bot.\n")
	message.WriteString("I'll inform you when a new issue appears on <b>SparkLink</b>.\n\n")
	message.WriteString("I'll focus on these tokens:\n\n")
	for _, tb := range binds {
		scanUrl, _ := tb.TokenScanURL()
		bindInfo := fmt.Sprintf(
			"<b>%s</b> (<a href=\"%s\">$%s</a>)\n",
			escapeHtml(tb.ERC20Name),
			scanUrl,
			escapeHtml(tb.ERC20Symbol),
		)
		message.WriteString(bindInfo)
	}

	err = m.EditOrReply(message.String(), &tele.SendOptions{
		DisableWebPagePreview:   true,
		ParseMode:               "html",
	})
	if err != nil {
		log.Errorf(err.Error())
	}
}

func sendErrorMessage(m tele.Context, message string, err error) {
	chatTitle, _ := getChatTitle(m.Chat().ID)
	m.EditOrReply(
		fmt.Sprintf("%s\\.\n\nPlease kick me off and invite me in again\\.", message),
	)
	log.Errorf(
		"%s. Group: %s (%d): %s",
		message,
		chatTitle,
		m.Chat().ID,
		err.Error(),
	)
}

func escapeHtml(text string) (result string) {
	result = strings.Replace(text, "&", "&amp;", -1)
	result = strings.Replace(result, "<", "&lt;", -1)
	result = strings.Replace(result, ">", "&rt;", -1)
	log.Debugf("Before: %s, After: %s", text, result)

	return result
}

package inline_keyboards

import (
	"github.com/calendar-bot/pkg/bots/telegram/messages"
	tb "gopkg.in/tucnak/telebot.v2"
)

const (
	contactUrl = "https://t.me/alexey_ershkov"
)

func ReportBugKeyboard() [][]tb.InlineButton {
	return [][]tb.InlineButton{{{
		Text: messages.GetMessageReportBug(),
		URL: contactUrl,
	}}}
}

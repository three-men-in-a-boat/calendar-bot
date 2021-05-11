package baseInlineKeyboards

import (
	"github.com/calendar-bot/pkg/bots/telegram/messages/baseMessages"
	tb "gopkg.in/tucnak/telebot.v2"
)

func StartInlineKeyboard(url string) [][]tb.InlineButton {
	inlineKeyboard := [][]tb.InlineButton{{
		{Text: baseMessages.StartRegButtonText(), URL: url},
	}}

	return inlineKeyboard
}

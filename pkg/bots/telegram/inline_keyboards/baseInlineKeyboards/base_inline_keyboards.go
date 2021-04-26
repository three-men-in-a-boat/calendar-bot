package baseInlineKeyboards

import (
	"github.com/calendar-bot/pkg/bots/telegram/messages/baseMessages"
	tb "gopkg.in/tucnak/telebot.v2"
)


func Start(url string) (keyboard *tb.ReplyMarkup) {
	keyboard = &tb.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(keyboard.URL(baseMessages.StartRegButtonText(), url)),
	)

	return keyboard
}

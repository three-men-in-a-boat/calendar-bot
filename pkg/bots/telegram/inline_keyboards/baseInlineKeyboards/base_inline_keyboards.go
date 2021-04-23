package baseInlineKeyboards

import (
	"github.com/calendar-bot/pkg/bots/telegram/text/baseText"
	tb "gopkg.in/tucnak/telebot.v2"
)


func Start(url string) (keyboard *tb.ReplyMarkup) {
	keyboard = &tb.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(keyboard.URL(baseText.StartRegButtonText(), url)),
	)

	return keyboard
}

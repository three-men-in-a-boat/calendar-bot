package inline_keyboards

import (
	"github.com/calendar-bot/pkg/bots/telegram/text"
	tb "gopkg.in/tucnak/telebot.v2"
)

func Start(url string) (keyboard *tb.ReplyMarkup) {
	keyboard = &tb.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(keyboard.URL(text.StartRegButtonText(), url)),
	)

	return keyboard
}

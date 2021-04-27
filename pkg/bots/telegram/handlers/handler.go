package handlers

import (
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Handler struct {
	bot *tb.Bot
}

func (h *Handler) SendError(sender tb.Recipient, err error) {
	h.bot.Send(sender, messages.MessageUnexpectedError(err.Error()),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
				InlineKeyboard: inline_keyboards.ReportBugKeyboard(),
			},
		})
}

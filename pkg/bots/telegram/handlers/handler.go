package handlers

import (
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages"
	"github.com/calendar-bot/pkg/customerrors"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Handler struct {
	bot *tb.Bot
	parseAddress string
}

func (h *Handler) SendError(sender tb.Recipient, outerErr error) {
	_, err := h.bot.Send(sender, messages.MessageUnexpectedError(outerErr.Error()),
		&tb.SendOptions{
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
				InlineKeyboard: inline_keyboards.ReportBugKeyboard(),
			},
		})

	if err != nil {
		customerrors.HandlerError(err)
	}
}

func (h *Handler) SendAuthError(sender tb.Recipient, outerErr error) {
	_, err := h.bot.Send(sender, messages.MessageAuthError(outerErr.Error()),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})

	if err != nil {
		customerrors.HandlerError(err)
	}
}

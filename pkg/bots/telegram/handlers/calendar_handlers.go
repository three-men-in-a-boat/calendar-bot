package handlers

import (
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/customerrors"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	tb "gopkg.in/tucnak/telebot.v2"
)

type CalendarHandlers struct {
	handler      Handler
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
}

func NewCalendarHandlers(eventUC eUseCase.EventUseCase, userUC uUseCase.UserUseCase) CalendarHandlers {
	return CalendarHandlers{eventUseCase: eventUC, userUseCase: userUC, handler: Handler{bot: nil}}
}

func (ch *CalendarHandlers) InitHandlers(bot *tb.Bot) {
	ch.handler.bot = bot
	bot.Handle("/today", ch.HandleToday)
}

func (ch *CalendarHandlers) HandleToday(m *tb.Message) {
	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	events, err := ch.eventUseCase.GetEventsToday(token)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	for _, event := range events.Data.Events {
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.SingleEventShortText(event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	}
}

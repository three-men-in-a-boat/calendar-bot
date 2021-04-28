package handlers

import (
	"context"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards/calendarInlineKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/customerrors"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

type CalendarHandlers struct {
	handler      Handler
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
	redisDB      *redis.Client
}

func NewCalendarHandlers(eventUC eUseCase.EventUseCase, userUC uUseCase.UserUseCase, redis *redis.Client) CalendarHandlers {
	return CalendarHandlers{eventUseCase: eventUC, userUseCase: userUC, handler: Handler{bot: nil}, redisDB: redis}
}

func (ch *CalendarHandlers) InitHandlers(bot *tb.Bot) {
	ch.handler.bot = bot
	bot.Handle("/today", ch.HandleToday)
	bot.Handle("\f"+telegram.ShowFullEvent, ch.HandleShowMore)
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
		keyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
		if err != nil {
			zap.S().Errorf("Can't set calendarId=%v for eventId=%v. Err: %v",
				event.Calendar.UID, event.Uid, err)
		}
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.SingleEventShortText(event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
				InlineKeyboard:      keyboard,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	}
}

func (ch *CalendarHandlers) HandleShowMore(c *tb.Callback) {
	calUid, err := ch.redisDB.Get(context.TODO(), c.Data).Result()
	if err != nil {
		customerrors.HandlerError(err)
		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text: calendarMessages.RedisNotFoundMessage(),
			ShowAlert: true,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return
	}

	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(c.Sender.ID))

	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendAuthError(c.Sender, err)

		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return
	}
	_, err = ch.eventUseCase.GetEventByEventID(token, calUid, c.Data)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(c.Sender, err)

		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return
	}

	err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text: "Заголовок события",
	})
	if err != nil {
		customerrors.HandlerError(err)
	}

	_, err = ch.handler.bot.Send(c.Sender, "Событие здесь")
	if err != nil {
		customerrors.HandlerError(err)
	}
}

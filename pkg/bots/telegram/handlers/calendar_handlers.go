package handlers

import (
	"context"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/inline_keyboards/calendarInlineKeyboards"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/customerrors"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/types"
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
	bot.Handle("/next", ch.HandleNext)
	bot.Handle("\f"+telegram.ShowFullEvent, ch.HandleShowMore)
	bot.Handle("\f"+telegram.ShowShortEvent, ch.HandleShowLess)
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
		keyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(&event, ch.redisDB)
		if err != nil {
			zap.S().Errorf("Can't set calendarId=%v for eventId=%v. Err: %v",
				event.Calendar.UID, event.Uid, err)
		}
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.SingleEventShortText(&event), &tb.SendOptions{
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

func (ch *CalendarHandlers) HandleNext(m *tb.Message) {
	token, err := ch.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(int64(m.Sender.ID))
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	event, err := ch.eventUseCase.GetClosestEvent(token)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(m.Sender, err)
		return
	}

	if event != nil {
		inlineKeyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
		if err != nil {
			customerrors.HandlerError(err)
		}
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.SingleEventShortText(event), &tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				ReplyKeyboardRemove: true,
				InlineKeyboard:      inlineKeyboard,
			},
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
	} else {
		_, err = ch.handler.bot.Send(m.Sender, calendarMessages.NoClosestEvents(), &tb.SendOptions{
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

func (ch *CalendarHandlers) getEventByIdForCallback(c *tb.Callback) *types.Event {
	calUid, err := ch.redisDB.Get(context.TODO(), c.Data).Result()
	if err != nil {
		customerrors.HandlerError(err)
		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
			Text:       calendarMessages.RedisNotFoundMessage(),
			ShowAlert:  true,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return nil
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
		return nil
	}
	resp, err := ch.eventUseCase.GetEventByEventID(token, calUid, c.Data)
	if err != nil {
		customerrors.HandlerError(err)
		ch.handler.SendError(c.Sender, err)

		err = ch.handler.bot.Respond(c, &tb.CallbackResponse{
			CallbackID: c.ID,
		})
		if err != nil {
			customerrors.HandlerError(err)
		}
		return nil
	}

	return &resp.Data.Event
}

func (ch *CalendarHandlers) HandleShowMore(c *tb.Callback) {

	event := ch.getEventByIdForCallback(c)
	if event == nil {
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
		Text:       calendarMessages.CallbackResponseHeader(event),
	})
	if err != nil {
		customerrors.HandlerError(err)
	}

	_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventFullText(event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: calendarInlineKeyboards.EventShowLessInlineKeyboard(event),
			},
		})
	if err != nil {
		customerrors.HandlerError(err)
	}
}

func (ch *CalendarHandlers) HandleShowLess(c *tb.Callback) {
	event := ch.getEventByIdForCallback(c)
	if event == nil {
		return
	}

	err := ch.handler.bot.Respond(c, &tb.CallbackResponse{
		CallbackID: c.ID,
	})
	if err != nil {
		customerrors.HandlerError(err)
	}

	inlineKeyboard, err := calendarInlineKeyboards.EventShowMoreInlineKeyboard(event, ch.redisDB)
	if err != nil {
		customerrors.HandlerError(err)
	}
	_, err = ch.handler.bot.Edit(c.Message, calendarMessages.SingleEventShortText(event),
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyMarkup: &tb.ReplyMarkup{
				InlineKeyboard: inlineKeyboard,
			},
		})
	if err != nil {
		customerrors.HandlerError(err)
	}
}

package calendarInlineKeyboards

import (
	"context"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/types"
	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
	"strings"
)

func EventShowMoreInlineKeyboard(event *types.Event, db *redis.Client) ([][]tb.InlineButton, error) {
	err := db.Set(context.TODO(), event.Uid, event.Calendar.UID, 0).Err()
	if err != nil {
		return nil, err
	}
	return [][]tb.InlineButton{{{
		Text:   calendarMessages.ShowMoreButton(),
		Unique: telegram.ShowFullEvent,
		Data:   event.Uid,
	}}}, nil
}

func EventShowLessInlineKeyboard(event *types.Event) [][]tb.InlineButton {
	inlineKeyboard := make([][]tb.InlineButton, 0)
	if event.Call != "" {
		inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{{
			Text: calendarMessages.CallLinkButton(),
			URL:  event.Call,
		}})
	}

	inlineKeyboard = append(inlineKeyboard, []tb.InlineButton{{
		Text: calendarMessages.ShowLessButton(),
		Unique: telegram.ShowShortEvent,
		Data: event.Uid,
	}})

	return inlineKeyboard
}

func GroupAlertsButtons(data string) [][]tb.InlineButton {
	inp := ""
	if strings.Contains(data, telegram.Today) {
		inp = telegram.Today
	}
	if strings.Contains(data, telegram.Next) {
		inp = telegram.Next
	}
	if strings.Contains(data, telegram.Date) {
		inp = telegram.Date
	}
	return [][]tb.InlineButton {{
		{
			Text: "Да",
			Unique: telegram.AlertCallbackYes,
			Data: inp,
		},
		{
			Text: "Нет",
			Unique: telegram.AlertCallbackNo,
		},
	}}
}

package calendarInlineKeyboards

import (
	"context"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/types"
	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
)

func EventShowMoreInlineKeyboard(event types.Event, db *redis.Client) ([][]tb.InlineButton, error) {
	err := db.Set(context.TODO(), event.Uid, event.Calendar.UID, 0).Err()
	if err != nil {
		return nil, err
	}
	return [][]tb.InlineButton{{{
		Text: calendarMessages.ShowMoreButton(),
		Unique: telegram.ShowFullEvent,
		Data: event.Uid,
	}}}, nil
}


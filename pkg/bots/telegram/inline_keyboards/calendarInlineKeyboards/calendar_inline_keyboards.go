package calendarInlineKeyboards

import (
	"context"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/types"
	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"strings"
	"time"
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
		Text:   calendarMessages.ShowLessButton(),
		Unique: telegram.ShowShortEvent,
		Data:   event.Uid,
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
	return [][]tb.InlineButton{{
		{
			Text:   "Да",
			Unique: telegram.AlertCallbackYes,
			Data:   inp,
		},
		{
			Text:   "Нет",
			Unique: telegram.AlertCallbackNo,
		},
	}}
}

func CreateEventButtons(event types.Event) [][]tb.InlineButton {
	btns := make([][]tb.InlineButton, 0)

	if !event.From.IsZero() && !event.To.IsZero() {
		btns = append(btns, []tb.InlineButton{{
			Text:   calendarMessages.GetCreateEventCreateText(),
			Unique: telegram.CreateEvent,
		}})
	}

	btns = append(btns, []tb.InlineButton{{
		Text:   calendarMessages.GetCreateCancelText(),
		Unique: telegram.CancelCreateEvent,
	}})

	return btns
}

func GroupChatButtons(event *types.Event, db *redis.Client, senderID int) ([][]tb.InlineButton, error) {
	err := db.Set(context.TODO(), event.Uid, event.Calendar.UID, 0).Err()
	if err != nil {
		return nil, err
	}
	return [][]tb.InlineButton{{
		{
			Text:   calendarMessages.CreateEventGo,
			Unique: telegram.GroupGo,
			Data:   event.Uid + "|" + strconv.Itoa(senderID),
		},
		{
			Text:   calendarMessages.CreateEventNotGo,
			Unique: telegram.GroupNotGo,
			Data:   event.Uid + "|" + strconv.Itoa(senderID),
		},
	}}, nil
}

func GroupFindTimeButtons() [][]tb.InlineButton {
	return [][]tb.InlineButton{{
		{
			Text:   calendarMessages.CreateEventFindTimeYesButton,
			Unique: telegram.GroupFindTimeYes,
		},
		{
			Text:   calendarMessages.CreateEventFindTimeNoButton,
			Unique: telegram.GroupFindTimeNo,
		},
	}}
}

func FindTimeDayPartButtons(t time.Time) [][]tb.InlineButton {
	return [][]tb.InlineButton{
		{
			{
				Text: "Утром",
				Unique: telegram.FindTimeDayPart,
				Data: time.Date(t.Year(), t.Month(), t.Day(), 6,0,0,0, t.Location()).Format(time.RFC3339),
			},
			{
				Text: "Днем",
				Unique: telegram.FindTimeDayPart,
				Data: time.Date(t.Year(), t.Month(), t.Day(), 12,0,0,0, t.Location()).Format(time.RFC3339),
			},
		},
		{
			{
				Text: "Вечером",
				Unique: telegram.FindTimeDayPart,
				Data: time.Date(t.Year(), t.Month(), t.Day(), 18,0,0,0, t.Location()).Format(time.RFC3339),
			},
			{
				Text: "Ночью",
				Unique: telegram.FindTimeDayPart,
				Data: time.Date(t.Year(), t.Month(), t.Day(), 0,0,0,0, t.Location()).Format(time.RFC3339),
			},
		},
		{
			{
				Text: "В любое время",
				Unique: telegram.FindTimeDayPart,
				Data: "All day",
			},

			{
				Text: calendarMessages.GetCreateCancelText(),
				Unique: telegram.FindTimeDayPart,
				Data: calendarMessages.GetCreateCancelText(),
			},
		},
	}
}

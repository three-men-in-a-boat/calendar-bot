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
				Text:   "Утром (6:00 - 13:00)",
				Unique: telegram.FindTimeDayPart,
				Data:   time.Date(t.Year(), t.Month(), t.Day(), 6, 0, 0, 0, t.Location()).Format(time.RFC3339),
			},
		},
		{
			{
				Text:   "Днем (12:00 - 19:00)",
				Unique: telegram.FindTimeDayPart,
				Data:   time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location()).Format(time.RFC3339),
			},
		},
		{
			{
				Text:   "Вечером (17:00 - 0:00)",
				Unique: telegram.FindTimeDayPart,
				Data:   time.Date(t.Year(), t.Month(), t.Day(), 17, 0, 0, 0, t.Location()).Format(time.RFC3339),
			},
		},
		{
			{
				Text:   "Ночью (0:00 - 7:00)",
				Unique: telegram.FindTimeDayPart,
				Data:   time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Format(time.RFC3339),
			},
		},
		{
			{
				Text:   "В любое время",
				Unique: telegram.FindTimeDayPart,
				Data:   "All day",
			},

			{
				Text:   calendarMessages.GetCreateCancelText(),
				Unique: telegram.FindTimeDayPart,
				Data:   calendarMessages.GetCreateCancelText(),
			},
		},
	}
}

func FindTimeLengthButtons() [][]tb.InlineButton {
	return [][]tb.InlineButton{
		{
			{
				Text:   "30 мин",
				Unique: telegram.FindTimeLength,
				Data:   "30m",
			},
			{
				Text:   "1 час",
				Unique: telegram.FindTimeLength,
				Data:   "1h",
			},
			{
				Text:   "1,5 часа",
				Unique: telegram.FindTimeLength,
				Data:   "1h30m",
			},
			{
				Text:   "2 часа",
				Unique: telegram.FindTimeLength,
				Data:   "2h",
			},
		},
		{
			{
				Text:   "2,5 часа",
				Unique: telegram.FindTimeLength,
				Data:   "2h30m",
			},
			{
				Text:   "3 часа",
				Unique: telegram.FindTimeLength,
				Data:   "3h",
			},
			{
				Text:   "4 часа",
				Unique: telegram.FindTimeLength,
				Data:   "4h",
			},
			{
				Text:   "5 часов",
				Unique: telegram.FindTimeLength,
				Data:   "5h",
			},
		},

		{
			{
				Text:   calendarMessages.GetCreateCancelText(),
				Unique: telegram.FindTimeLength,
				Data:   calendarMessages.GetCreateCancelText(),
			},
		},
	}
}

func FindTimePollButtons(sender int) [][]tb.InlineButton {
	return [][]tb.InlineButton{{
		{
			Text:   calendarMessages.FindTimeAdd,
			Unique: telegram.FindTimeAdd,
			Data:   strconv.Itoa(sender),
		},
	}, {
		{
			Text:   calendarMessages.GetCreateEventCreateText(),
			Unique: telegram.FindTimeCreate,
		},
		{
			Text:   calendarMessages.GetCreateCancelText(),
			Unique: telegram.FindTimeCreate,
			Data:   calendarMessages.GetCreateCancelText(),
		},
	},
	}
}

func GetDateFastCommand(cancelText bool) [][]tb.InlineButton {
	unique := telegram.HandleGroupText
	ret := [][]tb.InlineButton{
		{
			{
				Text:   "Сегодня",
				Unique: unique,
				Data:   "Сегодня",
			},
			{
				Text:   "Завтра",
				Unique: unique,
				Data:   "Завтра",
			},
			{
				Text:   "Через два дня",
				Unique: unique,
				Data:   "Через два дня",
			},
		},
		{
			{
				Text:   "Через три дня",
				Unique: unique,
				Data:   "Через три дня",
			},
			{
				Text:   "Через четыре дня",
				Unique: unique,
				Data:   "Через четыре дня",
			},
			{
				Text:   "Через пять дней",
				Unique: unique,
				Data:   "Через пять дней",
			},
		},
	}

	if !cancelText {
		ret = append(ret, []tb.InlineButton{
			{
				Text:   calendarMessages.GetCancelDateReplyButton(),
				Unique: unique,
				Data:   calendarMessages.GetCancelDateReplyButton(),
			},
		})
	} else {
		ret = append(ret, []tb.InlineButton{
			{
				Text:   calendarMessages.GetCreateCancelText(),
				Unique: unique,
				Data:   calendarMessages.GetCreateCancelText(),
			},
		})
	}

	return ret
}

func GetCreateFastCommand() [][]tb.InlineButton {
	unique := telegram.HandleGroupText
	return [][]tb.InlineButton{
		{
			{
				Text:   "Через полчаса",
				Unique: unique,
				Data:   "Через полчаса",
			},
			{
				Text:   "Через час",
				Unique: unique,
				Data:   "Через час",
			},
			{
				Text:   "Через два часа",
				Unique: unique,
				Data:   "Через два часа",
			},
			{
				Text:   "Через три часа",
				Unique: unique,
				Data:   "Через три часа",
			},
		},
		{
			{
				Text:   "Сегодня в 9:00",
				Unique: unique,
				Data:   "Сегодня в 9:00",
			},
			{
				Text:   "Сегодня в 12:00",
				Unique: unique,
				Data:   "Сегодня в 12:00",
			},
			{
				Text:   "Сегодня в 15:00",
				Unique: unique,
				Data:   "Сегодня в 15:00",
			},
			{
				Text:   "Сегодня в 18:00",
				Unique: unique,
				Data:   "Сегодня в 18:00",
			},
		},
		{
			{
				Text:   "Завтра в 9:00",
				Unique: unique,
				Data:   "Завтра в 9:00",
			},
			{
				Text:   "Завтра в 12:00",
				Unique: unique,
				Data:   "Завтра в 12:00",
			},
			{
				Text:   "Завтра в 15:00",
				Unique: unique,
				Data:   "Завтра в 15:00",
			},
			{
				Text:   "Завтра в 18:00",
				Unique: unique,
				Data:   "Завтра в 18:00",
			},
		},
		{
			{
				Text:   "Через неделю в 12:00",
				Unique: unique,
				Data:   "Через неделю в 12:00",
			},
			{
				Text:   "Через неделю в 15:00",
				Unique: unique,
				Data:   "Через неделю в 15:00",
			},
			{
				Text:   "Через неделю в 18:00",
				Unique: unique,
				Data:   "Через неделю в 18:00",
			},
		},
		{
			{
				Text:   calendarMessages.GetCreateCancelText(),
				Unique: unique,
				Data:   calendarMessages.GetCreateCancelText(),
			},
		},
	}
}

func GetCreateDuration() [][]tb.InlineButton {
	unique := telegram.HandleGroupText
	return [][]tb.InlineButton{
		{
			{
				Text:   calendarMessages.GetCreateEventHalfHour(),
				Unique: unique,
				Data:   calendarMessages.GetCreateEventHalfHour(),
			},
			{
				Text:   calendarMessages.GetCreateEventHour(),
				Unique: unique,
				Data:   calendarMessages.GetCreateEventHour(),
			},
			{
				Text:   calendarMessages.GetCreateEventHourAndHalf(),
				Unique: unique,
				Data:   calendarMessages.GetCreateEventHourAndHalf(),
			},
		},
		{
			{
				Text:   calendarMessages.GetCreateEventTwoHours(),
				Unique: unique,
				Data:   calendarMessages.GetCreateEventTwoHours(),
			},
			{
				Text:   calendarMessages.GetCreateEventFourHours(),
				Unique: unique,
				Data:   calendarMessages.GetCreateEventFourHours(),
			},
			{
				Text:   calendarMessages.GetCreateEventSixHours(),
				Unique: unique,
				Data:   calendarMessages.GetCreateEventSixHours(),
			},
		},
		{
			{
				Text:   calendarMessages.GetCreateFullDay(),
				Unique: unique,
				Data:   calendarMessages.GetCreateFullDay(),
			},
		},
	}
}

func GetCreateOptionButtons(session *types.BotRedisSession) [][]tb.InlineButton {
	btns := make([][]tb.InlineButton, 4)
	for i := range btns {
		btns[i] = make([]tb.InlineButton, 2)
	}
	idx := 0
	unique := telegram.HandleGroupText
	if session.Step != telegram.StepCreateFrom {
		btns[idx/2][idx%2] = tb.InlineButton{
			Text:   calendarMessages.CreateEventChangeStartTimeButton,
			Unique: unique,
			Data:   calendarMessages.CreateEventChangeStartTimeButton,
		}
		idx++
	}

	if session.Step != telegram.StepCreateTo {
		btns[idx/2][idx%2] = tb.InlineButton{
			Text:   calendarMessages.CreateEventChangeStopTimeButton,
			Unique: unique,
			Data:   calendarMessages.CreateEventChangeStopTimeButton,
		}
		idx++
	}

	if session.Step != telegram.StepCreateTitle {
		if session.Event.Title == "" {
			btns[idx/2][idx%2] = tb.InlineButton{
				Text:   calendarMessages.CreateEventAddTitleButton,
				Unique: unique,
				Data:   calendarMessages.CreateEventAddTitleButton,
			}
		} else {
			btns[idx/2][idx%2] = tb.InlineButton{
				Text:   calendarMessages.CreateEventChangeTitleButton,
				Unique: unique,
				Data:   calendarMessages.CreateEventChangeTitleButton,
			}
		}
		idx++
	}

	if session.Step != telegram.StepCreateDesc {
		if session.Event.Description == "" {
			btns[idx/2][idx%2] = tb.InlineButton{
				Text:   calendarMessages.CreateEventAddDescButton,
				Unique: unique,
				Data:   calendarMessages.CreateEventAddDescButton,
			}
		} else {
			btns[idx/2][idx%2] = tb.InlineButton{
				Text:   calendarMessages.CreateEventChangeDescButton,
				Unique: unique,
				Data:   calendarMessages.CreateEventChangeDescButton,
			}
		}
		idx++
	}

	if session.Step != telegram.StepCreateLocation {
		if session.Event.Location.Description == "" {
			btns[idx/2][idx%2] = tb.InlineButton{
				Text:   calendarMessages.CreateEventAddLocationButton,
				Unique: unique,
				Data:   calendarMessages.CreateEventAddLocationButton,
			}
		} else {
			btns[idx/2][idx%2] = tb.InlineButton{
				Text:   calendarMessages.CreateEventChangeLocationButton,
				Unique: unique,
				Data:   calendarMessages.CreateEventChangeLocationButton,
			}
		}
		idx++
	}

	if session.Step != telegram.StepCreateUser {
		btns[idx/2][idx%2] = tb.InlineButton{
			Text:   calendarMessages.CreateEventAddUser,
			Unique: unique,
			Data:   calendarMessages.CreateEventAddUser,
		}
		idx++
	}

	if !session.Event.FullDay {
		btns[idx/2][idx%2] = tb.InlineButton{
			Text:   calendarMessages.GetCreateFullDay(),
			Unique: unique,
			Data:   calendarMessages.GetCreateFullDay(),
		}
	}

	return btns
}

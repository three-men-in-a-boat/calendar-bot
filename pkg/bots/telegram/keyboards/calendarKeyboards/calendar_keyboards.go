package calendarKeyboards

import (
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	"github.com/calendar-bot/pkg/types"
	tb "gopkg.in/tucnak/telebot.v2"
)

func GetDateFastCommand(cancelText bool) [][]tb.ReplyButton {
	ret := [][]tb.ReplyButton{
		{
			{
				Text: "Сегодня",
			},
			{
				Text: "Завтра",
			},
			{
				Text: "Через два дня",
			},
		},
		{
			{
				Text: "Через неделю",
			},
			{
				Text: "Через две недели",
			},
			{
				Text: "Через месяц",
			},
		},
	}

	if !cancelText {
		ret = append(ret, []tb.ReplyButton{
			{
				Text: calendarMessages.GetCancelDateReplyButton(),
			},
		})
	} else {
		ret = append(ret, []tb.ReplyButton{
			{
				Text: calendarMessages.GetCreateCancelText(),
			},
		})
	}

	return ret
}

func GetCreateFastCommand() [][]tb.ReplyButton {
	return [][]tb.ReplyButton{
		{
			{
				Text: "Через полчаса",
			},
			{
				Text: "Через час",
			},
			{
				Text: "Через два часа",
			},
			{
				Text: "Через три часа",
			},
		},
		{
			{
				Text: "Сегодня в 9:00",
			},
			{
				Text: "Сегодня в 12:00",
			},
			{
				Text: "Сегодня в 15:00",
			},
			{
				Text: "Сегодня в 18:00",
			},
		},
		{
			{
				Text: "Завтра в 9:00",
			},
			{
				Text: "Завтра в 12:00",
			},
			{
				Text: "Завтра в 15:00",
			},
			{
				Text: "Завтра в 18:00",
			},
		},
		{
			{
				Text: "Через неделю в 12:00",
			},
			{
				Text: "Через неделю в 15:00",
			},
			{
				Text: "Через неделю в 18:00",
			},
		},
		{
			{
				Text: calendarMessages.GetCreateCancelText(),
			},
		},
	}
}

func GetCreateDuration() [][]tb.ReplyButton {
	return [][]tb.ReplyButton{
		{
			{
				Text: calendarMessages.GetCreateEventHalfHour(),
			},
			{
				Text: calendarMessages.GetCreateEventHour(),
			},
			{
				Text: calendarMessages.GetCreateEventHourAndHalf(),
			},
		},
		{
			{
				Text: calendarMessages.GetCreateEventTwoHours(),
			},
			{
				Text: calendarMessages.GetCreateEventFourHours(),
			},
			{
				Text: calendarMessages.GetCreateEventSixHours(),
			},
		},
		{
			{
				Text: calendarMessages.GetCreateFullDay(),
			},
		},
	}
}

func GetCreateOptionButtons(session *types.BotRedisSession) [][]tb.ReplyButton {
	btns := make([][]tb.ReplyButton, 4)
	for i := range btns {
		btns[i] = make([]tb.ReplyButton, 2)
	}
	idx := 0
	if session.Step != telegram.StepCreateFrom {
		btns[idx/2][idx%2] = tb.ReplyButton{
			Text: calendarMessages.CreateEventChangeStartTimeButton,
		}
		idx++
	}

	if session.Step != telegram.StepCreateTo {
		btns[idx/2][idx%2] = tb.ReplyButton{
			Text: calendarMessages.CreateEventChangeStopTimeButton,
		}
		idx++
	}

	if session.Step != telegram.StepCreateTitle {
		if session.Event.Title == "" {
			btns[idx/2][idx%2] = tb.ReplyButton{
				Text: calendarMessages.CreateEventAddTitleButton,
			}
		} else {
			btns[idx/2][idx%2] = tb.ReplyButton{
				Text: calendarMessages.CreateEventChangeTitleButton,
			}
		}
		idx++
	}

	if session.Step != telegram.StepCreateDesc {
		if session.Event.Description == "" {
			btns[idx/2][idx%2] = tb.ReplyButton{
				Text: calendarMessages.CreateEventAddDescButton,
			}
		} else {
			btns[idx/2][idx%2] = tb.ReplyButton{
				Text: calendarMessages.CreateEventChangeDescButton,
			}
		}
		idx++
	}

	if session.Step != telegram.StepCreateLocation {
		if session.Event.Location.Description == "" {
			btns[idx/2][idx%2] = tb.ReplyButton{
				Text: calendarMessages.CreateEventAddLocationButton,
			}
		} else {
			btns[idx/2][idx%2] = tb.ReplyButton{
				Text: calendarMessages.CreateEventChangeLocationButton,
			}
		}
		idx++
	}

	if session.Step != telegram.StepCreateUser {
		btns[idx/2][idx%2] = tb.ReplyButton{
			Text: calendarMessages.CreateEventAddUser,
		}
		idx++
	}

	if !session.Event.FullDay {
		btns[idx/2][idx%2] = tb.ReplyButton{
			Text: calendarMessages.GetCreateFullDay(),
		}
		idx++
	}

	return btns
}

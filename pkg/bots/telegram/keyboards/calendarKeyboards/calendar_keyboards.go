package calendarKeyboards

import (
	"github.com/calendar-bot/pkg/bots/telegram/messages/calendarMessages"
	tb "gopkg.in/tucnak/telebot.v2"
)

func GetDateFastCommand() [][]tb.ReplyButton {
	return [][]tb.ReplyButton{
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
		{
			{
				Text: calendarMessages.GetCancelDateReplyButton(),
			},
		},
	}
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

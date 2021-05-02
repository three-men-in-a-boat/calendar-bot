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
				Text: "Через неделю",
			},
		},
		{

			{
				Text: "Через две недели",
			},
			{
				Text: "В субботу",
			},
			{
				Text: "В воскресенье",
			},
		},
		{
			{
				Text: calendarMessages.GetCancelDateReplyButton(),
			},
		},
	}
}

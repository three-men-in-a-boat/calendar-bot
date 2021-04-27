package calendarMessages

import (
	"fmt"
	"github.com/calendar-bot/pkg/types"
	"github.com/goodsign/monday"
)

const (
	eventShortText = "<b>%s</b>\n\n⏰ %s, <u>%s</u> - %s<u>%s</u>\n" +
		"---------------\n" +
		"🗓 Календарь <b>%s</b>"
)

const (
	formatDate = "2 January 2006"
	formatTime = "15:04"
	locale = monday.LocaleRuRU
)

func parseDate(event types.Event) (from string, to string) {
	from = monday.Format(event.From, formatDate, locale)
	to = ""
	if event.From.Year() != event.To.Year() ||
		event.From.Month() != event.To.Month() ||
		event.From.Day() != event.To.Day() {
		to = monday.Format(event.To, formatDate, locale)
		to += ", "
	}

	return from, to
}

func SingleEventShortText(event types.Event) string {
	dateFrom, dateTo := parseDate(event)
	return fmt.Sprintf(
		eventShortText,
		event.Title,
		dateFrom,
		event.From.Format(formatTime),
		dateTo,
		event.To.Format(formatTime),
		event.Calendar.Title,
	)
}

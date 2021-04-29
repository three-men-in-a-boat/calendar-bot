package calendarMessages

import (
	"fmt"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/types"
	"github.com/goodsign/monday"
)

const (
	eventNameText                  = "<b>%s</b>\n\n"
	eventNoTitleText               = "Без названия"
	eventTimeText                  = "⏰ %s, <u>%s</u> - %s<u>%s</u>\n"
	eventTimeFullDay               = "⏰ %s %s, <u>Весь день</u>\n"
	eventPlaceText                 = "📍 %s\n"
	eventOrganizerText             = "Создатель - <b>%s</b> (%s)\n"
	eventSplitLine                 = "---------------\n"
	eventCalendarText              = "🗓 Календарь <b>%s</b>"
	eventAttendeesHeaderText       = "<u><i>Участники:</i></u>\n\n"
	eventAttendeeText              = "%s (%s) "
	eventAttendeeStatusAccepted    = "✅\n"
	eventAttendeeStatusNeedsAction = "❓\n"
	eventAttendeeStatusDeclined    = "❌\n"
	eventDescriptionHeader         = "<u><i>Описание:</i></u>\n\n"
	eventConfroomsHeader           = "<u><i>Переговорные комнаты:</i></u>\n\n"

	eventShowNotFoundError = "К сожалению мы не смогли найти информацию о событии.\n Возможно, это старое сообщение." +
		"\nЗапросите событие с помощью бота заново."
	eventCallbackResponseText = "Событие: %s"
)

const (
	formatDate = "2 January 2006"
	formatTime = "15:04"
	locale     = monday.LocaleRuRU
)

const (
	callLinkButton = "📲 Ссылка на звонок"
	showMoreButton = "🔻 Развернуть"
	showLessButton = "🔺 Свернуть"
)

func parseDate(event *types.Event) []interface{} {
	fromDate := monday.Format(event.From, formatDate, locale)
	toDate := ""
	if event.From.Year() != event.To.Year() ||
		event.From.Month() != event.To.Month() ||
		event.From.Day() != event.To.Day() {
		toDate = monday.Format(event.To, formatDate, locale)
		toDate += ", "
	}

	return []interface{}{fromDate, event.From.Format(formatTime), toDate, event.To.Format(formatTime)}
}

func parseDateFullDay(event *types.Event) []interface{} {
	fromDate := monday.Format(event.From, formatDate, locale)
	toDate := ""
	if event.From.Year() != event.To.Year() ||
		event.From.Month() != event.To.Month() ||
		event.From.Day()+1 != event.To.Day() {
		toDate += "- "
		toDate += monday.Format(event.To, formatDate, locale)
	}

	return []interface{}{fromDate, toDate}
}

func SingleEventShortText(event *types.Event) string {
	shortEventText := ""
	title := event.Title
	if title == "" {
		title = eventNoTitleText
	}
	shortEventText += fmt.Sprintf(eventNameText, title)
	if !event.FullDay {
		shortEventText += fmt.Sprintf(eventTimeText, parseDate(event)...)
	} else {
		shortEventText += fmt.Sprintf(eventTimeFullDay, parseDateFullDay(event)...)
	}
	shortEventText += eventSplitLine
	shortEventText += fmt.Sprintf(eventCalendarText, event.Calendar.Title)

	return shortEventText
}

func SingleEventFullText(event *types.Event) string {
	fullEventText := ""
	title := event.Title
	if title == "" {
		title = eventNoTitleText
	}
	fullEventText += fmt.Sprintf(eventNameText, title)
	if !event.FullDay {
		fullEventText += fmt.Sprintf(eventTimeText, parseDate(event)...)
	} else {
		fullEventText += fmt.Sprintf(eventTimeFullDay, parseDateFullDay(event)...)
	}
	if event.Location.Description != "" {
		fullEventText += fmt.Sprintf(eventPlaceText, event.Location.Description)
	}
	organizer := event.Organizer.Name
	if organizer != "" {
		organizer += " "
	}
	email := event.Organizer.Email
	if email == telegram.CalendarInternalEmail {
		organizer = telegram.MailRuCalendarName
		email = telegram.MailRuDomain
	}
	fullEventText += "\n" + fmt.Sprintf(eventOrganizerText, organizer, email)

	if len(event.Location.Confrooms) > 0 {
		fullEventText += eventSplitLine
		fullEventText += eventConfroomsHeader
		for _, confroom := range event.Location.Confrooms {
			fullEventText += confroom + "\n"
		}
	}

	if len(event.Attendees) > 1 {
		fullEventText += eventSplitLine
		fullEventText += eventAttendeesHeaderText
		for _, attendee := range event.Attendees {
			if attendee.Email == event.Organizer.Email {
				continue
			}
			fullEventText += fmt.Sprintf(eventAttendeeText, attendee.Name, attendee.Email)
			switch attendee.Status {
			case "ACCEPTED":
				fullEventText += eventAttendeeStatusAccepted
				break
			case "DECLINED":
				fullEventText += eventAttendeeStatusDeclined
				break
			default:
				fullEventText += eventAttendeeStatusNeedsAction
				break
			}
		}
	}

	if event.Description != "" {
		fullEventText += eventSplitLine
		fullEventText += eventDescriptionHeader
		fullEventText += event.Description + "\n"
	}

	fullEventText += eventSplitLine
	fullEventText += fmt.Sprintf(eventCalendarText, event.Calendar.Title)
	return fullEventText
}

func RedisNotFoundMessage() string {
	return eventShowNotFoundError
}

func ShowMoreButton() string {
	return showMoreButton
}

func ShowLessButton() string {
	return showLessButton
}

func CallLinkButton() string {
	return callLinkButton
}

func CallbackResponseHeader(event *types.Event) string {
	return fmt.Sprintf(eventCallbackResponseText, event.Title)
}

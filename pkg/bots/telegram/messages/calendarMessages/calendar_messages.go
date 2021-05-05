package calendarMessages

import (
	"fmt"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/types"
	"github.com/goodsign/monday"
	"strings"
	"time"
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

	eventTodayTitle = "<b>Ваши события на сегодня</b>"
	eventDateTitle  = "<b>Ваши события за %s</b>"
	eventNextTitle  = "<b>Ваше следуюшее событие</b>"

	eventGetDateHeader  = "<b>Получение событий за определенную дату: </b>\n\n"
	eventGetDateMessage = "Для выбора даты воспользуйтесь кнопками или введите дату в формате " +
		"<pre>&lt;число&gt; &lt;название месяца&gt;</pre> (например: <pre>22 марта</pre>)"

	eventNoTodayEventsFound  = "У вас нет событий сегодня"
	eventNoDateEventsFound   = "У вас нет событий за выбранную дату"
	eventNoClosestEventFound = "У вас больше нет событий сегодня"

	eventDateNotParsed   = "Мы не смогли распознать дату, попробуйте еще раз"
	eventSessionNotFound = "Мы не смогли найти необходимую информацию для обработки запроса.\nВоспользуйтесь нужной вам " +
		"командой заново"
	eventShowNotFoundError = "К сожалению мы не смогли найти информацию о событии.\nВозможно, это старое сообщение." +
		"\nЗапросите событие с помощью бота заново."
	eventCallbackResponseText = "Событие: %s"

	eventCancelSearchDate   = "Отмена поиска событий за опреденную дату"
	eventCanceledSearchDate = "Поиск события отменен"

	createEventHeader    = "<b>Что получается:</b>\n\n"
	createEventTitle     = "<b>Название:</b> %s\n"
	createEventAllDay    = "весь день"
	createEventDateStart = "⏰ <b>Начало:</b> %s %s\n"
	createEventDateEnd   = "⏰ <b>Конец:</b> %s %s\n"

	createEventInitText = "<b>Введите время начала события</b>\n\nДля выбора даты и времени начала события" +
		" воспользуйтесь кнопками или введите дату в формате <pre>&lt;число&gt; &lt;название месяца&gt; " +
		"&lt;ЧЧ:ММ&gt;</pre> (например: <pre>22 марта 15:00</pre>)"
	createEventCancelText   = "Отмена создания события"
	createEventCanceledText = "Создание события отменено"

	middlewaresUserNotAuthenticated = "Вы не можете воспользоваться данной функцией пока не авторизуетесь в боте через" +
		" аккаунт mail.ru. Для авторизации воспользуйтесь командой /start."
	middlewaresGroupAlertBase  = "Вы уверены, что хотите показать "
	middlewaresGroupAlertToday = "<b>ВСЕМ</b> свои события на сегодня?"
	middlewaresGroupAlertNext  = "<b>ВСЕМ</b> своё следующее событие на сегодня?"
	middlewaresGroupAlertDate  = "<b>ВСЕМ</b> свои события за определенную дату?"
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
func NoClosestEvents() string {
	return eventNoClosestEventFound
}

func GetInitDateMessage() string {
	return eventGetDateHeader + eventGetDateMessage
}

func GetTodayTitle() string {
	return eventTodayTitle
}

func GetDateTitle(date time.Time) string {
	return fmt.Sprintf(eventDateTitle, monday.Format(date, formatDate, locale))
}

func GetNextTitle() string {
	return eventNextTitle
}

func GetTodayNotFound() string {
	return eventNoTodayEventsFound
}

func GetDateEventsNotFound() string {
	return eventNoDateEventsFound
}

func GetDateNotParsed() string {
	return eventDateNotParsed
}

func RedisSessionNotFound() string {
	return eventSessionNotFound
}

func GetCancelDateReplyButton() string {
	return eventCancelSearchDate
}

func GetCancelDate() string {
	return eventCanceledSearchDate
}

func GetUserNotAuth() string {
	return middlewaresUserNotAuthenticated
}

func GetMessageAlertBase() string {
	return middlewaresGroupAlertBase
}

func GetGroupAlertMessage(data string) string {
	str := middlewaresGroupAlertBase
	if strings.Contains(data, telegram.Today) {
		return str + middlewaresGroupAlertToday
	}

	if strings.Contains(data, telegram.Next) {
		return str + middlewaresGroupAlertNext
	}

	if strings.Contains(data, telegram.Date) {
		return str + middlewaresGroupAlertDate
	}

	return ""
}

func GetCreateInitText() string {
	return createEventInitText
}

func GetCreateCancelText() string {
	return createEventCancelText
}

func GetCreateCanceledText() string {
	return createEventCanceledText
}

func GetCreateEventText(e *types.Event) string {
	str := ""
	str += createEventHeader

	if e.Title == "" {
		str += fmt.Sprintf(createEventTitle, eventNoTitleText)
	} else {
		str += fmt.Sprintf(createEventTitle, eventNoTitleText)
	}

	if !e.From.IsZero() {
		if !e.FullDay {
			str += fmt.Sprintf(createEventDateStart, monday.Format(e.From, formatDate, locale),
				monday.Format(e.From, formatTime, locale))
		} else {
			str += fmt.Sprintf(createEventDateStart, monday.Format(e.From, formatDate, locale),
				createEventAllDay)
		}
	}

	if !e.To.IsZero() && !e.FullDay {
		str += fmt.Sprintf(createEventDateEnd, monday.Format(e.From, formatDate, locale),
			monday.Format(e.From, formatTime, locale))
	}

	return str
}

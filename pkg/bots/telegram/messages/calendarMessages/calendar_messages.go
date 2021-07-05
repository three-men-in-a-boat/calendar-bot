package calendarMessages

import (
	"fmt"
	"github.com/calendar-bot/pkg/bots/telegram"
	"github.com/calendar-bot/pkg/types"
	"github.com/goodsign/monday"
	"github.com/senseyeio/spaniel"
	"strings"
	"time"
)

const (
	eventNameText                  = "<b>%s</b>\n\n"
	eventNoTitleText               = "Без названия"
	eventTimeText                  = "<u>Когда:</u>\n\n⏰ %s, <u>%s</u> - %s<u>%s</u>\n"
	eventTimeFullDay               = "<u>Когда:</u>\n\n⏰ %s %s, <u>Весь день</u>\n"
	eventDateStart                 = "<u>Когда:</u>\n\n⏰ <b>Начало:</b> %s %s\n"
	eventPlaceText                 = "\n<u>Где:</u>\n\n📍 %s\n"
	eventOrganizerText             = "\nСоздатель - <b>%s</b> (%s)\n"
	eventSplitLine                 = "---------------\n"
	EventCalendarText              = "🗓 Календарь <b>%s</b>"
	eventAttendeesHeaderText       = "<u><i>Участники:</i></u>\n\n"
	eventAttendeeText              = "%s (%s) "
	eventAttendeeStatusAccepted    = "✅\n"
	eventAttendeeStatusNeedsAction = "❓\n"
	eventAttendeeStatusDeclined    = "❌\n"
	eventDescriptionHeader         = "<u><i>Описание:</i></u>\n\n"
	eventConfroomsHeader           = "<u><i>Переговорные комнаты:</i></u>\n\n"

	eventTodayTitle = "<b>Ваши события на сегодня</b>"
	eventDateTitle  = "<b>Ваши события за %s</b>"
	eventNextTitle  = "<b>Ваше следующее событие</b>"

	eventGetDateHeader          = "<b>Получение событий за определенную дату: </b>\n\n"
	findTimeStartHeader         = "<b>Выберите дату начала для поиска времени: </b>\n\n"
	findTimeStopHeader          = "<b>Выберите дату оконачания для поиска времени: </b>\n\n"
	findTimeInfoHeader          = "<b>Поиск будет производиться этом временном промежутке </b>\n\n"
	findTimeStartTime           = "<b>Начало поиска: </b> %s\n"
	findTimeStopTime            = "<b>Окончание поиска: </b> %s"
	findTimeTextFormat          = "%s с %s до %s"
	FindTimeChooseDayPartHeader = "<b>Выберите период дня для события</b>\n\n"
	FindTimeChooseLengthHeader  = "<b>Выберите продолжительность события</b>\n\n"
	eventGetDateMessage         = "Для выбора даты воспользуйтесь кнопками или введите дату в формате " +
		"<pre>&lt;число&gt; &lt;название месяца&gt;</pre> (например: <pre>22 марта</pre>)"

	eventNoTodayEventsFound  = "У вас нет событий сегодня"
	eventNoDateEventsFound   = "У вас нет событий за выбранную дату"
	eventNoClosestEventFound = "У вас больше нет событий сегодня"

	findTimePollHeader = "Выберите время. Если хотите, чтобы ваш календарь учитывался - нажмите на кнопку " +
		"\"Участвую\"\n\nНа данный момент учитываются календари: "
	FindTimeAdd             = "✅ Участвую"
	FindTimeFind            = "Найти время"
	FindTimeBack            = "Изменить параметры"
	FindTimeExist           = "Вы уже участвуете"
	FindTimeNotFound        = "К сожалению мы не нашли свободного времени для всех участников с учетом данных параметров"
	FindTimePeriodIsTooLong = "Выбранный период для поиска слишком большой, сокращаем до максимального возможного"

	eventDateNotParsed      = "Мы не смогли распознать дату, попробуйте еще раз"
	EventDateToIsBeforeFrom = "<b>Введенная дата раньше начала события, введите корректную дату.</b>\n\n" +
		"Если вы хотите переставить события на другое время - измените время начала события"
	eventSessionNotFound = "Мы не смогли найти необходимую информацию для обработки запроса.\nВоспользуйтесь нужной вам " +
		"командой заново"
	EventNoEventDataFound = "<b>Мы не смогли распознать данные о событии. Попробуйте сформулировать предложение иначе." +
		"</b>\n\nНапример: Учеба завтра с 10:00 до 13:00"
	eventShowNotFoundError = "К сожалению мы не смогли найти информацию о событии.\nВозможно, это старое сообщение." +
		"\nЗапросите событие с помощью бота заново."
	eventCallbackResponseText = "Событие: %s"

	eventCancelSearchDate   = "Отмена поиска событий"
	eventCanceledSearchDate = "Поиск события отменен"

	createEventHeader  = "<u><b>Что получается:</b></u>\n\n"
	createdEventHeader = "<u><b>Событие создано:</b></u>\n\n"

	createEventInitText = "<b>Введите время начала события</b>\n\nДля выбора даты и времени начала события" +
		" воспользуйтесь кнопками или введите дату в формате <pre>&lt;число&gt; &lt;название месяца&gt; " +
		"&lt;ЧЧ:ММ&gt;</pre> (например: 22 марта 15:00)"
	createEventToText = "<b>Введите время окончания события</b>\n\nДля продолжительности события" +
		" воспользуйтесь кнопками или введите дату оконочания в формате <pre>&lt;число&gt; &lt;название месяца&gt; " +
		"&lt;ЧЧ:ММ&gt;</pre> (например: 22 марта 15:00)"
	createEventTitleText    = "<b>Введите название события</b>"
	CreateEventDescText     = "<b>Введите описание события</b>"
	CreateEventLocationText = "<b>Введите место события</b>"
	CreateEventUserText     = "<b>Введите email пользователя, которого хотите добавить. Можно ввести несколько" +
		" через запятую</b>"

	createEventCreateText   = "Создать событие"
	createEventCreatedText  = "Событие успешно создано"
	createEventCancelText   = "Отмена"
	createEventCanceledText = "Создание события отменено"

	createEventHalfHour    = "30 минут"
	createEventHour        = "1 час"
	createEventHourAndHalf = "1 час 30 минут"
	createEventTwoHours    = "2 часа"
	createEventFourHours   = "4 часа"
	createEventSixHours    = "6 часов"
	createEventFullDay     = "Весь день"

	CreateEventChangeStartTimeButton = "Изменить время начала"
	CreateEventChangeStopTimeButton  = "Изменить время окончания"
	CreateEventAddTitleButton        = "Добавить название"
	CreateEventChangeTitleButton     = "Изменить название"
	CreateEventAddDescButton         = "Добавить описание"
	CreateEventChangeDescButton      = "Изменить описание"
	CreateEventAddLocationButton     = "Добавить место"
	CreateEventChangeLocationButton  = "Изменить место"
	CreateEventAddUser               = "Добавить участников"

	ShowTodayTasks  = "покажи задачи на сегодня"
	ShowTodayPhrase = "покажи события на сегодня"
	ShowNextTask    = "покажи следующее событие"
	ShowNextPhrase  = "покажи ближайшее события"

	CreateEventGo    = "✅ Я иду"
	CreateEventNotGo = "❌ Я не иду"

	CreateEventAlreadyOrganize = "Вы являетесь создателем события и участуете в нем по умолчанию"
	CreateEventAlreadyGo       = "Вы уже дали согласие на участие в событии"
	CreateEventAlreadyNotGo    = "Вы уже дали отказ от участия в событии"
	CreateEventCannotAdd       = "Мы не смогли изменить ваш статус в событии - подтвердите его с помощью календаря или письма" +
		" из почты"

	CreateEventFindTimeMessage   = "Как создать событие?"
	CreateEventFindTimeYesButton = "🔍 Найти удобное время"
	CreateEventFindTimeNoButton  = "⏰ Создать на конкретное время"

	middlewaresUserNotAuthenticated = "Вы не можете воспользоваться данной функцией пока не авторизуетесь в боте через" +
		" аккаунт mail.ru. Для авторизации воспользуйтесь командой /start."
	middlewaresGroupAlertBase  = "Вы уверены, что хотите показать "
	middlewaresGroupAlertToday = "<b>ВСЕМ</b> свои события на сегодня?"
	middlewaresGroupAlertNext  = "<b>ВСЕМ</b> своё следующее событие на сегодня?"
	middlewaresGroupAlertDate  = "<b>ВСЕМ</b> свои события за определенную дату?"

	userNotAllow = "Вы не можете взаимодействовать с данной кнопкой"
)

const (
	formatSpan = "2 January"
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

func SingleEventShortText(event *types.Event, isCalendarSend bool) string {
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

	if isCalendarSend {
		shortEventText += eventSplitLine
		shortEventText += fmt.Sprintf(EventCalendarText, event.Calendar.Title)
	}
	return shortEventText
}

func SingleEventFullText(event *types.Event) string {
	fullEventText := ""
	title := event.Title
	if title == "" {
		title = eventNoTitleText
	}
	fullEventText += fmt.Sprintf(eventNameText, title)

	if event.To.IsZero() && !event.From.IsZero() {
		fullEventText += fmt.Sprintf(eventDateStart, monday.Format(event.From, formatDate, locale),
			monday.Format(event.From, formatTime, locale))
	} else {
		if !event.FullDay {
			fullEventText += fmt.Sprintf(eventTimeText, parseDate(event)...)
		} else {
			fullEventText += fmt.Sprintf(eventTimeFullDay, parseDateFullDay(event)...)
		}
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
		fullEventText += eventAttendeesHeaderText
		for _, attendee := range event.Attendees {
			if attendee.Email == event.Organizer.Email {
				continue
			}
			fullEventText += fmt.Sprintf(eventAttendeeText, attendee.Name, attendee.Email)
			switch attendee.Status {
			case telegram.StatusAccepted:
				fullEventText += eventAttendeeStatusAccepted
			case telegram.StatusDeclined:
				fullEventText += eventAttendeeStatusDeclined
			default:
				fullEventText += eventAttendeeStatusNeedsAction
			}
		}
	}

	if event.Description != "" {
		fullEventText += eventSplitLine
		fullEventText += eventDescriptionHeader
		fullEventText += event.Description + "\n"
	}

	if event.Calendar.Title != "" {
		fullEventText += eventSplitLine
		fullEventText += fmt.Sprintf(EventCalendarText, event.Calendar.Title)
	}

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

func GetCreateEventHeader() string {
	return createEventHeader
}

func GetCreateEventCreateText() string {
	return createEventCreateText
}

func GetCreateEventHalfHour() string {
	return createEventHalfHour
}
func GetCreateEventHour() string {
	return createEventHour
}
func GetCreateEventHourAndHalf() string {
	return createEventHourAndHalf
}
func GetCreateEventTwoHours() string {
	return createEventTwoHours
}
func GetCreateEventFourHours() string {
	return createEventFourHours
}
func GetCreateEventSixHours() string {
	return createEventSixHours
}

func GetCreateEventToText() string {
	return createEventToText
}

func GetUserNotAllow() string {
	return userNotAllow
}

func GetEventCreatedText() string {
	return createEventCreatedText
}

func GetCreatedEventHeader() string {
	return createdEventHeader
}

func GetCreateFullDay() string {
	return createEventFullDay
}

func GetCreateEventTitle() string {
	return createEventTitleText
}

func GetFindTimeStartText() string {
	return findTimeStartHeader + eventGetDateMessage
}

func GetFindTimeStopText(time time.Time) string {
	return findTimeStopHeader + fmt.Sprintf(findTimeStartTime, monday.Format(time, formatDate, locale)) +
		"\n" + eventGetDateMessage
}

func GetFindTimeInfoText(timeFrom, timeTo time.Time) string {
	return findTimeInfoHeader + fmt.Sprintf(findTimeStartTime, monday.Format(timeFrom, formatDate, locale)) +
		fmt.Sprintf(findTimeStopTime, monday.Format(timeTo, formatDate, locale))
}

func GenOptionsForPoll(spans spaniel.Spans) []string {
	str := make([]string, 0)
	counter := 0
	for _, span := range spans {
		if counter == 9 {
			return str
		}
		str = append(str, fmt.Sprintf(findTimeTextFormat,
			monday.Format(span.Start(), formatSpan, locale),
			monday.Format(span.Start(), formatTime, locale),
			monday.Format(span.End(), formatTime, locale),
		))
		counter++
	}

	return str
}

func GenFindTimePollHeader(emails []string) string {
	return findTimePollHeader + strings.Join(emails, ", ")
}

func AddNameBold(name string) string {
	return "<b>, " + name + "</b>"
}

func AddName(name string) string {
	return ", " + name
}

func AddNameStartBold(name string) string {
	return "<b>" + name + ": </b>"
}

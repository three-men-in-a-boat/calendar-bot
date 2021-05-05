package telegram

const (
	ShowFullEvent    = "SFE"
	ShowShortEvent   = "SSE"
	ShowGroupToday   = "SGT"
	ShowGroupNext    = "SGN"
	ShowGroupDate    = "SGD"
	AlertCallbackYes = "ACY"
	AlertCallbackNo  = "ACN"

	Today = "/today"
	Next  = "/next"
	Date  = "/date"

	CalendarInternalEmail = "calendar@internal"
	MailRuDomain          = "mail.ru"
	MailRuCalendarName    = "Календарь Mail"
)

const (
	StepCreateInit int = iota
	StepCreateFrom
	StepCreateTo
	StepCreateTitle
	StepCreateDesc
	StepCreateUser
)

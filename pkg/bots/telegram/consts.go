package telegram

const (
	ShowFullEvent     = "SFE"
	ShowShortEvent    = "SSE"
	CreateEvent       = "CRE"
	CancelCreateEvent = "CCE"
	AlertCallbackYes  = "ACY"
	AlertCallbackNo   = "ACN"
	GroupGo = "GG"
	GroupNotGo = "GNG"

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
	StepCreateLocation

	RoleRequired      = "REQUIRED"
	StatusNeedsAction = "NEEDS_ACTION"
	StatusAccepted    = "ACCEPTED"
	StatusDeclined    = "DECLINED"
)

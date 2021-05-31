package telegram

const (
	ShowFullEvent     = "SFE"
	ShowShortEvent    = "SSE"
	CreateEvent       = "CRE"
	CancelCreateEvent = "CCE"
	AlertCallbackYes  = "ACY"
	AlertCallbackNo   = "ACN"
	GroupGo           = "GG"
	GroupNotGo        = "GNG"
	GroupFindTimeYes  = "GFTY"
	GroupFindTimeNo   = "GFTN"
	FindTimeDayPart   = "FTDP"
	FindTimeLength    = "FTL"
	FindTimeAdd       = "FTA"
	FindTimeFind      = "FTF"
	FindTimeBack      = "FTB"
	FindTimeCreate    = "FTC"

	HandleGroupText = "HGT"

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

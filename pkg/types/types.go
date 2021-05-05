package types

import (
	"fmt"
	"github.com/calendar-bot/pkg/bots/telegram/utils"
	"time"
)

type StatesDictionary struct {
	States map[string]string
}

func NewStatesDictionary() StatesDictionary {
	return StatesDictionary{
		States: map[string]string{},
	}
}

type Calendar struct {
	UID   string `json:"uid,omitempty"`
	Title string `json:"title,omitempty"`
}

type AttendeeEvent struct {
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

type LocationEvent struct {
	Description string   `json:"description,omitempty"`
	Confrooms   []string `json:"confrooms,omitempty"`
	Geo         Geo      `json:"geo,omitempty"`
}

type Geo struct {
	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
}

type AttendeesEvent []AttendeeEvent

type Events []Event

type Event struct {
	Uid         string         `json:"uid,omitempty"`
	Title       string         `json:"title,omitempty"`
	From        time.Time      `json:"from,omitempty"`
	To          time.Time      `json:"to,omitempty"`
	FullDay     bool           `json:"fullDay,omitempty"`
	Description string         `json:"description,omitempty"`
	Location    LocationEvent  `json:"location,omitempty"`
	Calendar    Calendar       `json:"calendar,omitempty"`
	Attendees   AttendeesEvent `json:"attendees,omitempty"`
	Call        string         `json:"call,omitempty"`
	Organizer   AttendeeEvent  `json:"organizer,omitempty"`
	Payload     string         `json:"payload,omitempty"`
}

type DataEvents struct {
	Events Events `json:"events,omitempty"`
}

type EventsResponse struct {
	Data DataEvents `json:"data,omitempty"`
}

type DataEvent struct {
	Event Event `json:"event,omitempty"`
}

type EventResponse struct {
	Data DataEvent `json:"data,omitempty"`
}

type MailruAPIResponseErr struct {
	ErrorName        string `json:"error,omitempty"`
	ErrorCode        int    `json:"error_code,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (o *MailruAPIResponseErr) Error() string {
	if o == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"MailruAPIResponseErr: error='%s', error_code=%d, error_description='%s'",
		o.ErrorName, o.ErrorCode, o.ErrorDescription,
	)
}

func (o *MailruAPIResponseErr) IsError() bool {
	if o == nil {
		return false
	}
	if o.ErrorName != "" {
		return true
	}
	return false
}

func (o *MailruAPIResponseErr) GetError() *MailruAPIResponseErr {
	return o
}

type TelegramDBUser struct {
	ID               int64
	MailUserID       string
	MailUserEmail    string // nickeskov: maybe also remove this field?
	MailRefreshToken string
	TelegramUserId   int64
	CreatedAt        time.Time
}

// Gender: m - male, f - female

type MailruUserInfo struct {
	ID        string `json:"id"`
	Gender    string `json:"gender"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Locale    string `json:"locale"`
	Email     string `json:"email"`
	Birthday  string `json:"birthday"`
	Image     string `json:"image"`
	*MailruAPIResponseErr
	//ClientID  string
}

func (m *MailruUserInfo) IsValid() bool {
	if m == nil {
		return false
	}
	return !m.MailruAPIResponseErr.IsError()
}

type Location struct {
	Description string   `json:"description,omitempty"`
	Confrooms   []string `json:"confrooms,omitempty"`
	Geo         *Geo     `json:"geo,omitempty"`
}

type Attendee struct {
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}

type Attendees []Attendee

type EventInput struct {
	Uid         *string    `json:"uid,omitempty"`
	Title       *string    `json:"title,omitempty"`
	From        *string    `json:"from,omitempty"`
	To          *string    `json:"to,omitempty"`
	FullDay     *bool      `json:"fullDay,omitempty"`
	Description *string    `json:"description,omitempty"`
	Location    *Location  `json:"location,omitempty"`
	Calendar    *string    `json:"calendar,omitempty"`
	Attendees   *Attendees `json:"attendees,omitempty"`
	Call        *string    `json:"call,omitempty"`
	Chat        *string    `json:"chat,omitempty"`
	Payload     *string    `json:"payload,omitempty"`
}

type AddAttendee struct {
	EventID    string `json:"eventID,omitempty"`
	CalendarID string `json:"calendarID,omitempty"`
	Email      string `json:"email,omitempty"`
	Role       string `json:"role,omitempty"`
}

type ChangeStatus struct {
	EventID    string `json:"eventID,omitempty"`
	CalendarID string `json:"calendarID,omitempty"`
	Status     string `json:"status,omitempty"`
}

type FreeBusy struct {
	Users []string `json:"users,omitempty"`
	From  string   `json:"from,omitempty"`
	To    string   `json:"to,omitempty"`
}

type FromTo struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type FreeBusyIntervals struct {
	User     string   `json:"user,omitempty"`
	FreeBusy []FromTo `json:"freebusy,omitempty"`
}

type FreeBusyUser struct {
	FreeBusy []FreeBusyIntervals `json:"freebusy,omitempty"`
}

type FreeBusyResponse struct {
	Data FreeBusyUser `json:"data,omitempty"`
}


type BotRedisSession struct {
	Step int
	IsDate bool `json:"is_date"`
	IsCreate bool `json:"is_create"`
	Event Event
	InfoMsg utils.CustomEditable
}

type ParseDateReq struct {
	Timezone string `json:"timezone,omitempty"`
	Text string `json:"text"`
}

type ParseDateResp struct {
	Date time.Time `json:"date,omitempty"`
}

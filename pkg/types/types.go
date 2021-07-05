package types

import (
	"github.com/calendar-bot/pkg/bots/telegram/utils"
	"github.com/senseyeio/spaniel"
	"time"
)

type Calendar struct {
	UID   string `json:"uid,omitempty"`
	Title string `json:"title,omitempty"`
	Type  string `json:"type,omitempty"`
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

type TelegramDBUser struct {
	ID                   int64
	MailUserID           string
	MailUserEmail        string
	MailRefreshToken     string
	TelegramUserId       int64
	TelegramUserTimezone *string
	CreatedAt            time.Time
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
	Users []string  `json:"users,omitempty"`
	From  time.Time `json:"from,omitempty"`
	To    time.Time `json:"to,omitempty"`
}

type FromTo struct {
	From time.Time `json:"from,omitempty"`
	To   time.Time `json:"to,omitempty"`
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

func (ft FromTo) Start() time.Time {
	return ft.From
}

func (ft FromTo) StartType() spaniel.EndPointType {
	return spaniel.Closed
}

func (ft FromTo) End() time.Time {
	return ft.To
}

func (ft FromTo) EndType() spaniel.EndPointType {
	return spaniel.Open
}

type DayPart struct {
	Start    time.Time
	Duration time.Duration
}

type BotRedisSession struct {
	Step             int                  `json:"step"`
	FromTextCreate   bool                 `json:"from_text_create"`
	IsDate           bool                 `json:"is_date"`
	IsCreate         bool                 `json:"is_create"`
	FindTimeDone     bool                 `json:"find_time_done"`
	Event            Event                `json:"event"`
	FreeBusy         FreeBusy             `json:"free_busy"`
	FindTimeDayPart  *DayPart             `json:"day_part"`
	FindTimeDuration time.Duration        `json:"find_time_duration"`
	Users            []int64              `json:"users"`
	InfoMsg          utils.CustomEditable `json:"info_msg"`
	PollMsg          utils.CustomEditable `json:"poll_msg"`
	InlineMsg        utils.CustomEditable `json:"inline_msg"`
	FindTimeInfoMsg  utils.CustomEditable `json:"find_time_info_msg"`
}

type ParseDateReq struct {
	Timezone string `json:"timezone,omitempty"`
	Text     string `json:"text"`
}

type ParseDateResp struct {
	Date time.Time `json:"date,omitempty"`
}

type ParseEventResp struct {
	EventStart time.Time `json:"event_start,omitempty"`
	EventEnd   time.Time `json:"event_end,omitempty"`
	EventName  string    `json:"event_name,omitempty"`
}

type CreateEvent struct {
	Uid      string   `json:"uid,omitempty"`
	Calendar Calendar `json:"calendar,omitempty"`
}

type RespData struct {
	CreateEvent CreateEvent `json:"createEvent,omitempty"`
}

type CreateEventResp struct {
	Data RespData `json:"data,omitempty"`
}

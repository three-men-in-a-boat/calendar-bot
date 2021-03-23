package types

import (
	"fmt"
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
	ID    string `json:"uid,omitempty"`
	Title string `json:"title,omitempty"`
}

type Attendee struct {
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

type Attendees []Attendee

type Events []Event

type Event struct {
	Title       string    `json:"title,omitempty"`
	From        string    `json:"from,omitempty"`
	To          string    `json:"to,omitempty"`
	FullDay     bool      `json:"fullDay,omitempty"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Calendar    Calendar  `json:"calendar,omitempty"`
	Attendees   Attendees `json:"attendees,omitempty"`
	Call        string    `json:"call,omitempty"`
	Organizer   Attendee  `json:"organizer,omitempty"`
	Payload     string    `json:"payload,omitempty"`
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

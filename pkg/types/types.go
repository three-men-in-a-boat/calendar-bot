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

type Event struct {
	Name         string   `json:"name,omitempty"`
	Participants []string `json:"participants,omitempty"`
	Time         string   `json:"time,omitempty"`
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
	ID     int64
	UserID string // TODO(nickeskov): rename to MailUserID (in DB too)

	MailUserEmail    string // nickeskov: maybe also remove this field?
	MailRefreshToken string

	MailAccessToken    string    // TODO(nickeskov): remove this field
	MailTokenExpiresIn time.Time // TODO(nickeskov): remove this field

	TelegramUserId int64

	CreatedAt time.Time
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

package types

import "time"

type StatesDictionary struct {
	States map[int64]string
}

func NewStatesDictionary() StatesDictionary {
	return StatesDictionary{
		States: map[int64]string{},
	}
}

type Event struct {
	Name         string   `json:"name,omitempty"`
	Participants []string `json:"participants,omitempty"`
	Time         string   `json:"time,omitempty"`
}

type User struct {
	ID     int64
	UserID string

	MailUserEmail string

	MailAccessToken    string
	MailRefreshToken   string
	MailTokenExpiresIn time.Time

	TelegramUserId int64

	CreatedAt time.Time
}

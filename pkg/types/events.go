package types

type Event struct {
	Name string `json:"name,omitempty"`
	Participants []string `json:"participants,omitempty"`
	Time string `json:"time,omitempty"`
}
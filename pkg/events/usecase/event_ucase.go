package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/calendar-bot/pkg/events/repository"
	"github.com/calendar-bot/pkg/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

type EventUseCase struct {
	eventStorage repository.EventRepository
}

func NewEventUseCase(eventStor repository.EventRepository) EventUseCase {
	return EventUseCase{
		eventStorage: eventStor,
	}
}

func getStartDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Now().Location())
}
func getEndDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, time.Now().Location())
}

func sortEvents(events []types.Event) (types.Events, error) {
	sort.Slice(events, func(i, j int) bool {
		return events[i].From.Unix() < events[j].From.Unix()
	})
	return events, nil
}

func closestEvent(events []types.Event) (*types.Event, error) {
	for _, event := range events {
		if event.From.Unix() > time.Now().Unix() {
			return &event, nil
		}
	}
	return nil, nil
}

func getEventsBySpecificDay(t time.Time, accessToken string) (*types.EventsResponse, error) {
	startDay := getStartDay(t).Format(time.RFC3339)
	endDay := getEndDay(t).Format(time.RFC3339)

	graphqlRequest := fmt.Sprintf(`
	{
		events(from: "%s", to: "%s") {
			uid,
			title,
			from,
			to,
			fullDay,
			description,
			location{
				description,
				confrooms,
				geo {
     				latitude,
					longitude,
				}
 			},
			calendar {
 				uid,
 				title
			},
 			attendees{
				email,
   			name,
   			role,
   			status
 			},
 			call,
 			organizer{
   			email,
   			name,
   			role,
   			status
 			},
 			payload
  		}
	}
	`, startDay, endDay)

	request, err := http.NewRequest("GET", "https://calendar.mail.ru/graphql", nil)
	if err != nil {
		zap.S().Errorf("failed to send request %s", err)
		return nil, err
	}
	q := request.URL.Query()
	q.Add("query", graphqlRequest)
	request.URL.RawQuery = q.Encode()

	var bearerToken = "Bearer " + accessToken
	request.Header.Add(
		"Authorization",
		bearerToken,
	)

	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Errorf("The HTTP request failed with error %v", err)
	}
	defer func() {
		err = response.Body.Close()
		if err != nil {
			zap.S().Errorf("failed to close body of response of func getEvents, %v", err)
		}
	}()

	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	eventsResponse := types.EventsResponse{}

	err = json.Unmarshal(res, &eventsResponse)
	if err != nil {
		return nil, err
	}
	if len(eventsResponse.Data.Events) == 0 {
		return nil, nil
	}

	events, err := sortEvents(eventsResponse.Data.Events)
	if err != nil {
		return nil, err
	}
	eventsResponse.Data.Events = events

	return &eventsResponse, nil
}

func (uc *EventUseCase) GetEventsToday(accessToken string) (*types.EventsResponse, error) {
	return getEventsBySpecificDay(time.Now(), accessToken)
}

func (uc *EventUseCase) GetClosestEvent(accessToken string) (*types.Event, error) {
	eventsResponse, err := uc.GetEventsToday(accessToken)
	if err != nil {
		return nil, err
	}

	closestEvent, err := closestEvent(eventsResponse.Data.Events)
	if err != nil {
		return nil, err
	}
	return closestEvent, nil
}

func (uc *EventUseCase) GetEventsByDate(accessToken string, date time.Time) (*types.EventsResponse, error) {
	return getEventsBySpecificDay(date, accessToken)
}

type EventInput struct {
	Uid         string    `json:"uid,omitempty"`
	Title       string    `json:"title,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	FullDay     bool      `json:"fullDay,omitempty"`
	Description string    `json:"description,omitempty"`
	Location    types.Location  `json:"location,omitempty"`
	Calendar    types.Calendar  `json:"calendar,omitempty"`
	Attendees   types.Attendees `json:"attendees,omitempty"`
	Call        string    `json:"call,omitempty"`
	Chat		string 		`json:"chat,omitempty"`
	Payload     string    `json:"payload,omitempty"`
}

type HTTPResponse struct {
	statusCode int
	response string
}

func (uc *EventUseCase) CreateEvent(accessToken string, eventInput EventInput) (*HTTPResponse, error) {
	mutationReq := fmt.Sprintf(`mutation{createEvent(event: {uid: \"%s\", title: \"%s\", from: \"%s\", to: \"%s\", description: \"%s\"}) {uid}}`, eventInput.Uid, eventInput.Title, eventInput.From, eventInput.To,eventInput.Description)
	eventCreationReq := fmt.Sprintf(`{"query":"%s"}`, mutationReq)

	request, err := http.NewRequest("POST", "https://calendar.mail.ru/graphql", bytes.NewBuffer([]byte(eventCreationReq)))
	if err != nil {
		zap.S().Errorf("failed to create a request %s", err)
		return nil, err
	}

	var bearerToken = "Bearer " + accessToken
	request.Header.Add(
		"Authorization",
		bearerToken,
	)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Errorf("The HTTP request failed with error %v", err)
	}
	defer func() {
		err = response.Body.Close()
		if err != nil {
			zap.S().Errorf("failed to close body of response of func getEvents, %v", err)
		}
	}()

	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	eventsResponse := types.EventsResponse{}

	err = json.Unmarshal(res, &eventsResponse)
	if err != nil {
		return nil, err
	}
	responseHTTP := HTTPResponse{}
	responseHTTP.statusCode = response.StatusCode
	responseHTTP.response = string(res)

	return &responseHTTP, nil
}


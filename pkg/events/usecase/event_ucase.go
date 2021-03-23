package usecase

import (
	"fmt"
	"github.com/calendar-bot/pkg/events/repository"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
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
	return time.Date(year, month, day, 0, 0, 0, 0, &time.Location{})
}
func getEndDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, &time.Location{})
}

func (uc *EventUseCase) GetEventsToday(accessToken string) (*string, error) {
	startDay := getStartDay(time.Now()).Format(time.RFC3339)
	endDay := getEndDay(time.Now()).Format(time.RFC3339)

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

	events, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	res := string(events)
	return &res, nil
}

package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/calendar-bot/pkg/events/repository"
	"github.com/calendar-bot/pkg/types"
	"github.com/fatih/structs"
	"github.com/pkg/errors"
	"github.com/senseyeio/spaniel"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
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

func closestEvent(events []types.Event) *types.Event {
	if events == nil {
		return nil
	}

	min := events[0]

	for _, event := range events {
		if event.To.Unix() > time.Now().Unix() && event.From.Unix() < min.From.Unix() {
			min = event
		}
	}
	if min.To.Unix() > time.Now().Unix() {
		return &min
	}
	return nil
}

func getEventsBySpecificDay(t time.Time, accessToken string) (*types.EventsResponse, error) {
	startDay := getStartDay(t).Format(time.RFC3339)
	endDay := getEndDay(t).Format(time.RFC3339)

	graphqlRequest := fmt.Sprintf(`
	{
		events(from: "%s", to: "%s", buildVirtual: true) {
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

	// sort slice in order by time
	sort.Slice(eventsResponse.Data.Events, func(i, j int) bool {
		return eventsResponse.Data.Events[i].From.Unix() < eventsResponse.Data.Events[j].From.Unix()
	})

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
	if eventsResponse == nil {
		return nil, nil
	}

	closestEvent := closestEvent(eventsResponse.Data.Events)

	return closestEvent, nil
}

func (uc *EventUseCase) GetEventsByDate(accessToken string, date time.Time) (*types.EventsResponse, error) {
	return getEventsBySpecificDay(date, accessToken)
}

func (uc *EventUseCase) GetEventByEventID(accessToken string, calendarID string, eventID string) (*types.EventResponse, error) {
	graphqlRequest := fmt.Sprintf(`
	{
		event(eventUID: "%s", calendarUID: "%s") {
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
	`, eventID, calendarID)

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
	eventResponse := types.EventResponse{}

	err = json.Unmarshal(res, &eventResponse)
	if err != nil {
		return nil, err
	}

	return &eventResponse, nil
}

func (uc *EventUseCase) GetUsersBusyIntervals(accessToken string, freeBusy types.FreeBusy) (*types.FreeBusyResponse, error) {
	var users string
	for i, user := range freeBusy.Users {
		users += "\"" + user + "\""
		if i != len(freeBusy.Users)-1 {
			users += ","
		}
	}
	graphqlRequest := fmt.Sprintf(
		`{freebusy(from: "%s", to: "%s", forUsers: [%s]) {user, freebusy{from, to}}}`,
		freeBusy.From.Format(time.RFC3339),
		freeBusy.To.Format(time.RFC3339),
		users,
	)

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

	freeBusyResponse := types.FreeBusyResponse{}

	err = json.Unmarshal(res, &freeBusyResponse)
	if err != nil {
		return nil, err
	}

	return &freeBusyResponse, nil
}

func (uc *EventUseCase) GetUsersFreeIntervals(accessToken string, freeBusy types.FreeBusy,
	conf FreeBusyConfig) (spaniel.Spans, error) {

	response, err := uc.GetUsersBusyIntervals(accessToken, freeBusy)
	if err != nil {
		return nil, errors.Wrap(err, "GetUsersFreeIntervals")
	}

	freeBusyBorders := spaniel.New(freeBusy.From, freeBusy.To)
	busyFlatTimeSpan := MergeBusyIntervals(response.Data, conf.StretchBusyIntervalsBy)

	busyFlatTruncated := MapSpansWithFunc(busyFlatTimeSpan, TruncateSpanBy(freeBusyBorders))

	freeTimeSpans := CalculateFreeTimeSpans(busyFlatTruncated, freeBusyBorders)

	if conf.SplitFreeIntervalsBy != nil {
		splitBy := *conf.SplitFreeIntervalsBy
		freeTimeSplit := make(spaniel.Spans, 0, len(freeTimeSpans))
		for _, span := range freeTimeSpans {
			// neskov: denotes remainder
			spanSplit, _ := SplitSpanBy(span, splitBy)
			freeTimeSplit = append(freeTimeSplit, spanSplit...)
		}
		freeTimeSpans = freeTimeSplit
	}

	filteredFreeTimeSpans := FilterSpans(
		freeTimeSpans,
		nil,
		conf.DayPart,
		conf.MinFreeIntervalDuration,
		conf.MaxFreeIntervalDuration,
	)

	return filteredFreeTimeSpans, nil
}

func getNewTime(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Now().Location())
}

func RemoveLastChar(str string) string {
	for len(str) > 0 {
		_, size := utf8.DecodeLastRuneInString(str)
		return str[:len(str)-size]
	}
	return str
}

func MakeFirstLowerCase(s string) string {

	if len(s) < 2 {
		return strings.ToLower(s)
	}

	bts := []byte(s)

	lc := bytes.ToLower([]byte{bts[0]})
	rest := bts[1:]

	return string(bytes.Join([][]byte{lc, rest}, nil))
}

func getJsonFromMap(m map[string]interface{}) string {
	var response string
	for key, element := range m {
		key = MakeFirstLowerCase(key)
		switch value := element.(type) {
		case string:
			response += fmt.Sprintf("%s: \\\"%s\\\",", key, value)
		case *string:
			if value == nil {
				continue
			}
			response += fmt.Sprintf("%s: \\\"%s\\\",", key, *value)
		case []string:
			var array string
			for i, v := range value {
				array += fmt.Sprintf("\\\"%s\\\"", v)
				if i != len(value)-1 {
					array += ", "
				}
			}
			response += fmt.Sprintf("%s: [%s],", key, array)
		case *bool:
			if value == nil {
				continue
			}
			response += key + ": " + fmt.Sprintf("%t", *value) + ","
		case bool:
			response += key + ": " + fmt.Sprintf("%t", value) + ","
		case *map[string]interface{}:
			if value == nil {
				continue
			}
			lol := getJsonFromMap(*value)
			response += fmt.Sprintf("%s:{%s},", key, lol)
		case map[string]interface{}:
			lol := getJsonFromMap(value)
			response += fmt.Sprintf("%s:{%s},", key, lol)
		case *types.Attendees:
			if value == nil {
				continue
			}
			var array string

			for i, v := range *value {
				array += fmt.Sprintf("{email: \\\"%s\\\", role: %s}", v.Email, v.Role)
				if i != len(*value)-1 {
					array += ", "
				}
			}
			response += fmt.Sprintf("%s:[%s],", key, array)
		default:
			continue
		}

	}

	return response
}

func (uc *EventUseCase) CreateEvent(accessToken string, eventInput types.EventInput) ([]byte, error) {
	tmp, err := time.Parse(time.RFC3339, *eventInput.From)
	if err != nil {
		return nil, errors.Errorf("failed to parse `from` time, %v", err)
	}
	from := getNewTime(tmp).Format(time.RFC3339)
	eventInput.From = &from

	tmp, err = time.Parse(time.RFC3339, *eventInput.To)
	if err != nil {
		return nil, errors.Errorf("failed to parse `to` time, %v", err)
	}
	to := getNewTime(tmp).Format(time.RFC3339)
	eventInput.To = &to

	m := structs.Map(eventInput)

	queryEvent := getJsonFromMap(m)
	queryEvent = strings.ReplaceAll(queryEvent, ",}", "}")
	queryEvent = RemoveLastChar(queryEvent)

	mutationReq := fmt.Sprintf(`mutation{createEvent(event: {%s}) {uid,calendar{uid, title}}}`, queryEvent)
	eventCreationReq := fmt.Sprintf(`{"query":"%s"}`, mutationReq)

	request, err := http.NewRequest("POST", "https://calendar.mail.ru/graphql", bytes.NewBuffer([]byte(eventCreationReq)))
	if err != nil {
		return nil, errors.Errorf("failed to create a request: , %v", err)
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
			fmt.Printf("failed to close body of response of func getEvents, %v", err)
		}
	}()

	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read body %v", err)
	}
	return res, nil
}

func (uc *EventUseCase) AddAttendee(accessToken string, attendee types.AddAttendee) ([]byte, error) {
	mutationReq := fmt.Sprintf(`mutation{appendAttendee(uri: {uid: \"%s\", calendar: \"%s\"}, input: {attendee: {email: \"%s\", role: %s}}){email, role, name, status}}`, attendee.EventID, attendee.CalendarID, attendee.Email, attendee.Role)
	addAtttendeeRequest := fmt.Sprintf(`{"query":"%s"}`, mutationReq)

	request, err := http.NewRequest("POST", "https://calendar.mail.ru/graphql", bytes.NewBuffer([]byte(addAtttendeeRequest)))
	if err != nil {
		return nil, errors.Errorf("failed to create a request: , %v", err)
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
			fmt.Printf("failed to close body of response of func getEvents, %v", err)
		}
	}()

	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read body %v", err)
	}

	return res, nil
}

func (uc *EventUseCase) ChangeStatus(accessToken string, reactEvent types.ChangeStatus) ([]byte, error) {

	mutationReq := fmt.Sprintf(`mutation{reactEvent(input: {uid: \"%s\", calendar: \"%s\", status: %s}){uid,title, from, to, status}}`, reactEvent.EventID, reactEvent.CalendarID, reactEvent.Status)
	reactEventRequest := fmt.Sprintf(`{"query":"%s"}`, mutationReq)

	request, err := http.NewRequest("POST", "https://calendar.mail.ru/graphql", bytes.NewBuffer([]byte(reactEventRequest)))
	if err != nil {
		return nil, errors.Errorf("failed to create a request: , %v", err)
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
			fmt.Printf("failed to close body of response of func getEvents, %v", err)
		}
	}()

	res, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Errorf("failed to read body %v", err)
	}

	return res, nil
}

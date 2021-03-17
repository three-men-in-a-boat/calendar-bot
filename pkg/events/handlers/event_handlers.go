package handlers

import (
	"encoding/json"
	"github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/types"
	"github.com/labstack/echo"
	"net/http"
)

type EventHandlers struct {
	eventUseCase usecase.EventUseCase
}

func NewEventHandlers(eventUseCase usecase.EventUseCase) EventHandlers {
	return EventHandlers{eventUseCase: eventUseCase}
}

func (eh *EventHandlers) InitHandlers(server *echo.Echo) {

	server.GET("/api/v1/oauth/telegram/events", eh.getEvents)

}

func (eh *EventHandlers) getEvents(ctx echo.Context) error {
	var events []types.Event
	event1 := types.Event{Name: "Meeting in Zoom", Participants: []string{"Nikolay, Alexey, Alexandr"}, Time: "Today 23:00"}
	event2 := types.Event{Name: "Meeting in university", Participants: []string{"Mike, Alex, Gabe"}, Time: "Tomorrow 23:00"}
	events = append(events, event1, event2)
	b, err := json.Marshal(events)
	if err != nil {
		return err
	}
	return ctx.String(http.StatusOK, string(b))
}

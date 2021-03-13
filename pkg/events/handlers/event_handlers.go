package handlers

import (
	"github.com/calendar-bot/pkg/events/usecase"
	"github.com/labstack/echo"
)

type EventHandlers struct {
	eventUseCase usecase.EventUseCase
}

func NewEventHandlers(eventUseCase usecase.EventUseCase) EventHandlers {
	return EventHandlers{eventUseCase: eventUseCase}
}

func (e *EventHandlers) getEvents(rwContext echo.Context) error {

	return nil
}

func (e *EventHandlers) InitHandlers(server *echo.Echo) {

	server.GET("api/v1/getEvents", e.getEvents)

}

package handlers

import (
	"encoding/json"
	"fmt"
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/middlewares"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

type EventHandlers struct {
	eventUseCase eUseCase.EventUseCase
	userUseCase  uUseCase.UserUseCase
}

func NewEventHandlers(eventUseCase eUseCase.EventUseCase, userUseCase uUseCase.UserUseCase) EventHandlers {
	return EventHandlers{eventUseCase: eventUseCase, userUseCase: userUseCase}
}

func (eh *EventHandlers) InitHandlers(server *echo.Echo) {
	oauthMiddleware := middlewares.NewCheckOAuthTelegramMiddleware(&eh.userUseCase)

	eventRouter := server.Group("/api/v1/telegram/user/" + middlewares.TelegramUserIDRouteKey)

	eventRouter.GET("/events/today", eh.getEventsToday, oauthMiddleware.Handle)
	eventRouter.GET("/events/closest", eh.getClosesEvent, oauthMiddleware.Handle)
	eventRouter.GET("/events/date/:date", eh.getEventsByDate, oauthMiddleware.Handle)

	eventRouter.GET("/events/calendar/event", eh.getEventByEventID, oauthMiddleware.Handle)
}

func (eh *EventHandlers) getEventsToday(ctx echo.Context) error {
	telegramID, err := middlewares.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := middlewares.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	todaysEvent, err := eh.eventUseCase.GetEventsToday(accessToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get today's events for telegramUserID=%d", telegramID)
	}
	if todaysEvent == nil {
		return errors.New("response from calendar api for events is empty")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *todaysEvent)
}

func (eh *EventHandlers) getClosesEvent(ctx echo.Context) error {
	telegramID, err := middlewares.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := middlewares.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	closesEvent, err := eh.eventUseCase.GetClosesEvent(accessToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get the closest event for telegramUserID=%d", telegramID)
	}
	if closesEvent == nil {
		return errors.New("response from calendar api for events is empty")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *closesEvent)
}

func (eh *EventHandlers) getEventsByDate(ctx echo.Context) error {
	telegramID, err := middlewares.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := middlewares.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	dateFromCtx := ctx.Param("date")
	date, err := time.Parse(time.RFC3339, dateFromCtx)
	if err != nil {
		return err
	}

	eventsByDate, err := eh.eventUseCase.GetEventsByDate(accessToken, date)
	if err != nil {
		return errors.Wrapf(err, "failed to get the closest event for telegramUserID=%d", telegramID)
	}
	if eventsByDate == nil {
		return errors.New("response from calendar api for events is empty")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *eventsByDate)
}

type EventCalendarIDs struct {
	CalendarID string `json:"calendar_id,omitempty"`
	EventID    string `json:"event_id,omitempty"`
}

func (eh *EventHandlers) getEventByEventID(ctx echo.Context) error {
	telegramID, err := middlewares.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := middlewares.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	eventCalendarIDs := EventCalendarIDs{}

	b, err := ioutil.ReadAll(ctx.Request().Body)
	defer func() {
		err := ctx.Request().Body.Close()
		if err != nil {
			zap.S().Errorf("failed to close body %s", err)
		}
	}()

	if err != nil {
		return errors.Errorf("failed to read content from body")
	}
	err = json.Unmarshal(b, &eventCalendarIDs)
	if err != nil {
		return errors.Errorf("failed to unmarshal content from body")
	}

	event, err := eh.eventUseCase.GetEventByEventID(accessToken, eventCalendarIDs.CalendarID, eventCalendarIDs.EventID)
	if err != nil {
		return errors.Wrapf(err, "failed to get event by event_id and calendar_id for telegramUserID=%d", telegramID)
	}
	if event == nil {
		return ctx.String(http.StatusNotFound, fmt.Sprintf("event by event_id %s and calendar_id %s is not found for telegram id %d", eventCalendarIDs.EventID, eventCalendarIDs.CalendarID, telegramID))
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *event)
}

package handlers

import (
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/middlewares"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
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
	eventRouter.GET("/events/closest", eh.getClosestEvent, oauthMiddleware.Handle)
	eventRouter.GET("/events/date/:date", eh.getEventsByDate, oauthMiddleware.Handle)
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
		return ctx.String(http.StatusNotFound, "no events")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *todaysEvent)
}

func (eh *EventHandlers) getClosestEvent(ctx echo.Context) error {
	telegramID, err := middlewares.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := middlewares.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	closesEvent, err := eh.eventUseCase.GetClosestEvent(accessToken)
	if err != nil {
		return errors.Wrapf(err, "failed to get the closest event for telegramUserID=%d", telegramID)
	}
	if closesEvent == nil {
		return ctx.String(http.StatusNotFound, "no events")
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
		return ctx.String(http.StatusNotFound, "no events")
	}
	ctx.Response().Header().Set("Content-Type", "application/json")

	return ctx.JSON(http.StatusOK, *eventsByDate)
}

package handlers

import (
	eUseCase "github.com/calendar-bot/pkg/events/usecase"
	"github.com/calendar-bot/pkg/middlewares"
	uUseCase "github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
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

	userRouter := server.Group("/api/v1/telegram/user/" + middlewares.TelegramUserIDRouteKey)

	userRouter.GET("/events/today", eh.getEventsToday, oauthMiddleware.Handle)
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
		return errors.Wrapf(err, "failed to get userinfo for telegramUserID=%d", telegramID)
	}
	if todaysEvent == nil {
		return errors.New("response from calendar api for events is empty")
	}

	return ctx.String(http.StatusOK, *todaysEvent)
}

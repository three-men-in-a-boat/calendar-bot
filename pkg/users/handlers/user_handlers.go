package handlers

import (
"github.com/calendar-bot/pkg/users/usecase"
"github.com/labstack/echo"
)

type UserHandlers struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandlers(eventUseCase usecase.UserUseCase) UserHandlers {
	return UserHandlers{userUseCase: eventUseCase}
}

func (e *UserHandlers) getEvents(rwContext echo.Context) error {

	return nil
}

func (e *UserHandlers) InitHandlers(server *echo.Echo) {
	server.GET("api/v1/getEvents", e.getEvents)
}

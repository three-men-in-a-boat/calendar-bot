package handlers

import (
	"github.com/calendar-bot/pkg/statesDict"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"math/rand"
	"net/http"
	"strconv"
)

type UserHandlers struct {
	userUseCase usecase.UserUseCase
	statesDict statesDict.StatesDictionary
}

func NewUserHandlers(eventUseCase usecase.UserUseCase, states statesDict.StatesDictionary) UserHandlers {
	return UserHandlers{userUseCase: eventUseCase, statesDict: states}
}

func (e *UserHandlers) changeStateInLink(c echo.Context) error {
	name := c.Param("name")
	seed, err := strconv.Atoi(name)
	if err != nil {
		return err
	}
	rand.Seed(int64(seed))
	state := rand.Int()


	e.statesDict.States[name] = state

	link := "https://oauth.mail.ru/xlogin?client_id=885a013d102b40c7a46a994bc49e68f1&response_type=code&scope=&redirect_uri=https://calendarbot.xyz/api/v1/oauth&state=" + strconv.Itoa(state)

	return c.String(http.StatusOK, link)
}

func (e *UserHandlers) InitHandlers(server *echo.Echo) {
	server.GET("/api/v1/oauth/telegram/:name/ref", e.changeStateInLink)
}

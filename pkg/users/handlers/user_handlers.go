package handlers

import (
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"math/rand"
	"net/http"
	"strconv"
)

const RedirectUrlProd = "https://t.me/three_man_in_boat_bot"
const linkForGeneratingState = "https://oauth.mail.ru/xlogin?client_id=885a013d102b40c7a46a994bc49e68f1&response_type=code&scope=&redirect_uri=https://calendarbot.xyz/api/v1/oauth&state="

type UserHandlers struct {
	userUseCase usecase.UserUseCase
	statesDict  types.StatesDictionary
}

func NewUserHandlers(eventUseCase usecase.UserUseCase, states types.StatesDictionary) UserHandlers {
	return UserHandlers{userUseCase: eventUseCase, statesDict: states}
}

func (uh *UserHandlers) InitHandlers(server *echo.Echo) {
	server.GET("/api/v1/oauth/telegram/:id/ref", uh.changeStateInLink)
	server.GET("/api/v1/oauth", uh.telegramOauth)
}

func (uh *UserHandlers) changeStateInLink(ctx echo.Context) error {
	id := ctx.Param("id")
	seed, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	rand.Seed(int64(seed))
	state := rand.Int()

	uh.statesDict.States[int64(state)] = id

	link := linkForGeneratingState + strconv.Itoa(state)

	return ctx.String(http.StatusOK, link)
}

func (uh *UserHandlers) telegramOauth(ctx echo.Context) error {
	values := ctx.Request().URL.Query()

	code := values.Get("code")
	state := values.Get("state")

	stateInt, err := strconv.Atoi(state)
	if err != nil {
		println(err.Error())
		return err
	}

	tgId, err := strconv.Atoi(uh.statesDict.States[int64(stateInt)])
	if err != nil {
		println(err.Error())
		return err
	}

	if err := uh.userUseCase.TelegramCreateUser(int64(tgId), code); err != nil {
		return err
	}

	if err := ctx.Redirect(http.StatusTemporaryRedirect, RedirectUrlProd); err != nil {
		return err
	}

	return nil
}

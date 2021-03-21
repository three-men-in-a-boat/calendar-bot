package handlers

import (
	"crypto/rand"
	"fmt"
	"github.com/calendar-bot/cmd/config"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

const (
	nonceBitsLength = 256
	nonceBase       = 16
)

type UserHandlers struct {
	userUseCase usecase.UserUseCase
	statesDict  types.StatesDictionary
	conf        *config.App
}

func NewUserHandlers(eventUseCase usecase.UserUseCase, states types.StatesDictionary, conf *config.App) UserHandlers {
	return UserHandlers{
		userUseCase: eventUseCase,
		statesDict:  states,
		conf:        conf,
	}
}

func (uh *UserHandlers) InitHandlers(server *echo.Echo) {
	server.GET("/api/v1/oauth/telegram/:id/ref", uh.changeStateInLink)
	server.GET("/api/v1/oauth", uh.telegramOauth)
}

func (uh *UserHandlers) changeStateInLink(ctx echo.Context) error {
	// TODO(nickeskov): check type of id == int64
	id := ctx.Param("id")

	bigInt, err := rand.Prime(rand.Reader, nonceBitsLength)
	if err != nil {
		return errors.WithStack(err)
	}

	state := bigInt.Text(nonceBase)

	uh.statesDict.States[state] = id

	link := uh.generateOAuthLink(state)

	return ctx.String(http.StatusOK, link)
}

func (uh *UserHandlers) telegramOauth(ctx echo.Context) error {
	values := ctx.Request().URL.Query()

	// TODO(nickeskov): check that parameters
	code := values.Get("code")
	state := values.Get("state")

	id, ok := uh.statesDict.States[state]
	if !ok {
		return fmt.Errorf("cannot find state=%s in states dictionary", state)
	}

	tgId, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	if err := uh.userUseCase.TelegramCreateUser(int64(tgId), code); err != nil {
		return err
	}

	if err := ctx.Redirect(http.StatusTemporaryRedirect, uh.conf.OAuth.TelegramBotRedirectURI); err != nil {
		return err
	}

	return nil
}

func (uh *UserHandlers) generateOAuthLink(state string) string {
	return fmt.Sprintf(
		"https://oauth.mail.ru/login?client_id=%s&response_type=code&scope=%s&redirect_uri=%s&state=%s",
		uh.conf.OAuth.ClientID,
		uh.conf.OAuth.Scope,
		uh.conf.OAuth.RedirectURI,
		state,
	)
}

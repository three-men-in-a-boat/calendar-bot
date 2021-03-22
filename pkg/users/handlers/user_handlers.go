package handlers

import (
	"github.com/calendar-bot/cmd/config"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
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
	server.GET("/api/v1/oauth/telegram/user/:id/ref", uh.changeStateInLink)
	server.GET("/api/v1/oauth", uh.telegramOauth)
}

func (uh *UserHandlers) changeStateInLink(ctx echo.Context) error {
	// TODO(nickeskov): check type of id == int64
	id := ctx.Param("id")

	telegramID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		const status = http.StatusBadRequest
		return ctx.String(status, http.StatusText(status))
	}

	link, err := uh.userUseCase.GenOauthLinkForTelegramID(telegramID)
	if err != nil {
		return errors.WithStack(err)
	}

	return ctx.String(http.StatusOK, link)
}

func (uh *UserHandlers) telegramOauth(ctx echo.Context) error {
	// TODO(nickeskov): add middleware for token renewal
	values := ctx.Request().URL.Query()

	// TODO(nickeskov): check that parameters
	code := values.Get("code")
	state := values.Get("state")

	if code == "" || state == "" {
		const status = http.StatusBadRequest
		return ctx.String(status, http.StatusText(status))
	}

	telegramID, err := uh.userUseCase.GetTelegramIDByState(state)
	switch {
	case err == repository.StateDoesNotExist:
		const status = http.StatusForbidden
		return ctx.String(status, http.StatusText(status))
	case err != nil:
		return errors.WithStack(err)
	}

	if err := uh.userUseCase.TelegramCreateUser(telegramID, code); err != nil {
		return errors.WithStack(err)
	}

	if err := ctx.Redirect(http.StatusTemporaryRedirect, uh.conf.OAuth.TelegramBotRedirectURI); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

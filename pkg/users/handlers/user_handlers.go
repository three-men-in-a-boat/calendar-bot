package handlers

import (
	"github.com/calendar-bot/cmd/config"
	"github.com/calendar-bot/pkg/middlewares"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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
	userRouter := server.Group("/api/v1/oauth/telegram/user/" + middlewares.TelegramUserIDRouteKey)

	userRouter.GET("/ref", uh.generateOAuthLinkWithState)
	userRouter.GET("/is_auth", uh.chekAuthOfTelegramUser)

	server.GET("/api/v1/oauth", uh.telegramOauth)
}

func (uh *UserHandlers) generateOAuthLinkWithState(ctx echo.Context) error {
	id := ctx.Param(middlewares.TelegramUserIDPathParamKey)

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

func (uh *UserHandlers) chekAuthOfTelegramUser(ctx echo.Context) error {
	id := ctx.Param(middlewares.TelegramUserIDPathParamKey)

	telegramID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		const status = http.StatusBadRequest
		return ctx.String(status, http.StatusText(status))
	}

	isAuth, err := uh.userUseCase.IsUserAuthenticatedByTelegramUserID(telegramID)
	if err != nil {
		return errors.Wrapf(err, "error in chekAuthOfTelegramUser handler")
	}

	var status int
	if isAuth {
		status = http.StatusOK
	} else {
		status = http.StatusUnauthorized
	}

	return ctx.String(status, http.StatusText(status))
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
		zap.S().Debugf("state='%s' does not exist in redis, maybe user already authenticated", state)
		return ctx.Redirect(http.StatusTemporaryRedirect, uh.conf.OAuth.TelegramBotRedirectURI)
	case err != nil:
		return errors.WithStack(err)
	}

	isAuthenticated, err := uh.userUseCase.IsUserAuthenticatedByTelegramUserID(telegramID)
	if err != nil {
		return errors.Wrapf(err, "cannot check user authentication for telegramUserID=%d", telegramID)
	}
	if !isAuthenticated {
		if err := uh.userUseCase.TelegramCreateAuthentificatedUser(telegramID, code); err != nil {
			return errors.Wrapf(err, "cannot create and authetificate user wit telegramUserID=%d", telegramID)
		}
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, uh.conf.OAuth.TelegramBotRedirectURI)
}

package handlers

import (
	"github.com/calendar-bot/cmd/config"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/calendar-bot/pkg/utils/contextutils"
	"github.com/calendar-bot/pkg/utils/pathutils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
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
	//oauthMiddleware := middlewares.NewCheckOAuthTelegramMiddleware(&uh.userUseCase).Handle
	//
	//userRouter := server.Group("/api/v1/oauth/telegram/user/" + middlewares.TelegramUserIDRouteKey)
	//
	//userRouter.GET("/ref", uh.generateOAuthLinkWithState)
	//userRouter.GET("/is_auth", uh.chekAuthOfTelegramUser)
	//userRouter.GET("/info", uh.getMailruUserInfo, oauthMiddleware)
	//
	//userRouter.DELETE("", uh.deleteLocalAuthenticatedUser)

	server.GET("/api/v1/oauth", uh.telegramOAuth)
}

func (uh *UserHandlers) generateOAuthLinkWithState(ctx echo.Context) error {
	telegramID, err := pathutils.GetTelegramUserIDFromPathParams(ctx)
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
	telegramID, err := pathutils.GetTelegramUserIDFromPathParams(ctx)
	if err != nil {
		const status = http.StatusBadRequest
		return ctx.String(status, http.StatusText(status))
	}

	isAuth, err := uh.userUseCase.IsUserAuthenticatedByTelegramUserID(telegramID)
	if err != nil {
		return errors.Wrapf(err, "cannot check oauth for telegramUserID=%d", telegramID)
	}

	var status int
	if isAuth {
		status = http.StatusOK
	} else {
		status = http.StatusUnauthorized
	}

	return ctx.String(status, http.StatusText(status))
}

func (uh *UserHandlers) telegramOAuth(ctx echo.Context) error {
	values := ctx.Request().URL.Query()

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
		if err := uh.userUseCase.TelegramCreateAuthenticatedUser(telegramID, code); err != nil {
			return errors.Wrapf(err, "cannot create and authetificate user wit telegramUserID=%d", telegramID)
		}
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, uh.conf.OAuth.TelegramBotRedirectURI)
}

func (uh *UserHandlers) getMailruUserInfo(ctx echo.Context) error {
	telegramID, err := contextutils.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := contextutils.GetOAuthAccessTokenFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	userInfo, err := uh.userUseCase.GetMailruUserInfo(accessToken)
	if err != nil {
		return errors.Wrapf(err, "cannot get userinfo for telegramUserID=%d", telegramID)
	}

	if !userInfo.IsValid() {
		return errors.Wrapf(userInfo.GetError(),
			"mailru oauth userinfo API, response error, telegramUserID=%d", telegramID)
	}

	return ctx.JSON(http.StatusOK, userInfo)
}

func (uh *UserHandlers) deleteLocalAuthenticatedUser(ctx echo.Context) error {
	telegramID, err := pathutils.GetTelegramUserIDFromPathParams(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := uh.userUseCase.DeleteLocalAuthenticatedUserByTelegramUserID(telegramID); err != nil {
		if _, ok := err.(repository.UserEntityError); ok {
			const status = http.StatusNoContent
			return ctx.String(status, http.StatusText(status))
		}
		return errors.Wrapf(err, "failed to delete local authenticated user by telegramUserID=%d", telegramID)
	}

	const status = http.StatusOK
	return ctx.String(status, http.StatusText(status))
}

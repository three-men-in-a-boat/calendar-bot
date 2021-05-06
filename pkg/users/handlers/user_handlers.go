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
	telegramID, err := middlewares.GetTelegramUserIDFromPathParams(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	link, err := uh.userUseCase.GenOauthLinkForTelegramID(telegramID)
	if err != nil {
		return errors.WithStack(err)
	}

	return ctx.String(http.StatusOK, link)
}

func (uh *UserHandlers) chekAuthOfTelegramUser(ctx echo.Context) error {
	telegramID, err := middlewares.GetTelegramUserIDFromPathParams(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	isAuth, err := uh.userUseCase.IsUserAuthenticatedByTelegramUserID(telegramID)
	if err != nil {
		return errors.Wrapf(err, "cannot check oauth for telegramUserID=%d", telegramID)
	}

	if isAuth {
		return ctx.String(http.StatusOK, http.StatusText(http.StatusOK))
	}
	return ctx.String(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
}

func (uh *UserHandlers) telegramOAuth(ctx echo.Context) error {
	values := ctx.Request().URL.Query()

	code := values.Get("code")
	state := values.Get("state")

	if code == "" || state == "" {
		return ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
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
	telegramID, err := middlewares.GetTelegramUserIDFromContext(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	accessToken, err := middlewares.GetOAuthAccessTokenFromContext(ctx)
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
	telegramID, err := middlewares.GetTelegramUserIDFromPathParams(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := uh.userUseCase.DeleteLocalAuthenticatedUserByTelegramUserID(telegramID); err != nil {
		if _, ok := err.(repository.UserEntityError); ok {
			return ctx.String(http.StatusNoContent, http.StatusText(http.StatusNoContent))
		}
		return errors.Wrapf(err, "failed to delete local authenticated user by telegramUserID=%d", telegramID)
	}

	return ctx.String(http.StatusOK, http.StatusText(http.StatusOK))
}

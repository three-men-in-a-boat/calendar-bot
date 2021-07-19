package handlers

import (
	"github.com/calendar-bot/pkg/services/oauth"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
)

type UserHandlers struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandlers(eventUseCase usecase.UserUseCase) UserHandlers {
	return UserHandlers{
		userUseCase: eventUseCase,
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

func (uh *UserHandlers) telegramOAuth(ctx echo.Context) error {
	values := ctx.Request().URL.Query()

	code := values.Get("code")
	state := values.Get("state")

	if code == "" || state == "" {
		return ctx.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	telegramID, err := uh.userUseCase.GetTelegramUserIDByState(state)
	switch {
	case err == oauth.StateKeyDoesNotExist:
		zap.S().Debugf("state='%s' does not exist in redis, maybe user already authenticated", state)
		return ctx.Redirect(http.StatusTemporaryRedirect, uh.userUseCase.GetTelegramBotRedirectURI())
	case err != nil:
		return errors.WithStack(err)
	}

	isAuthenticated, err := uh.userUseCase.IsUserAuthenticatedByTelegramUserID(telegramID)
	if err != nil {
		return errors.Wrapf(err, "cannot check user authentication for telegramUserID=%d", telegramID)
	}
	if !isAuthenticated {
		if err := uh.userUseCase.TelegramCreateAuthenticatedUser(telegramID, code); err != nil {
			return errors.Wrapf(err, "cannot create and autheticated user with telegramUserID=%d", telegramID)
		}
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, uh.userUseCase.GetTelegramBotRedirectURI())
}

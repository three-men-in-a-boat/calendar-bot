package middlewares

import (
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

const (
	TelegramUserIDRouteKey     = ":telegramUserID"
	TelegramUserIDPathParamKey = "telegramUserID"
	TelegramUserIDContextKey   = "#telegramUserID#"
)

type CheckOAuthTelegramMiddleware struct {
	userUseCase *usecase.UserUseCase
}

func NewCheckOAuthTelegramMiddleware(useCase *usecase.UserUseCase) CheckOAuthTelegramMiddleware {
	return CheckOAuthTelegramMiddleware{
		userUseCase: useCase,
	}
}

func (m CheckOAuthTelegramMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {
		telegramIDPathParam := context.Param(TelegramUserIDPathParamKey)

		telegramID, err := strconv.ParseInt(telegramIDPathParam, 10, 64)
		if err != nil {
			const status = http.StatusBadRequest
			return context.String(status, http.StatusText(status))
		}

		oAuthToken, err := m.userUseCase.GetOAuthTokenByTelegramID(telegramID)
		switch {
		case err == repository.OAuthAccessTokenDoesNotExist:
			oAuthToken, err = m.userUseCase.RefreshOAuthTokenByTelegramUserID(telegramID)
			if err != nil {
				switch concreteErr := err.(type) {
				case repository.OAuthError, repository.UserEntityError:
					const status = http.StatusForbidden
					return context.String(status, http.StatusText(status))
				default:
					return errors.WithStack(concreteErr)
				}
			}
		case err != nil:
			return errors.WithStack(err)
		}

		context.Set(TelegramUserIDContextKey, oAuthToken)

		return next(context)
	}
}

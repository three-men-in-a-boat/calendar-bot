package middlewares

import (
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/calendar-bot/pkg/utils/contextutils"
	"github.com/calendar-bot/pkg/utils/pathutils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"net/http"
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
		telegramID, err := pathutils.GetTelegramUserIDFromPathParams(context)
		if err != nil {
			return context.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}

		oAuthToken, err := m.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(telegramID)
		if err != nil {
			switch concreteErr := err.(type) {
			case repository.OAuthError, repository.UserEntityError:
				return context.String(http.StatusForbidden, http.StatusText(http.StatusForbidden))
			default:
				return errors.WithStack(concreteErr)
			}
		}

		context.Set(contextutils.OAuthAccessTokenContextKey, oAuthToken)
		context.Set(contextutils.TelegramUserIDContextKey, telegramID)

		return next(context)
	}
}

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
	OAuthAccessTokenContextKey = "#oauthAccessToken#"
	TelegramUserIDContextKey   = "#telegramUserID#"
)

func GetOAuthAccessTokenFromContext(ctx echo.Context) (string, error) {
	accessToken, ok := ctx.Get(OAuthAccessTokenContextKey).(string)
	if !ok {
		return "", errors.New("cannot get from echo.Context OAuthAccessToken")
	}

	return accessToken, nil
}

func GetTelegramUserIDFromContext(ctx echo.Context) (int64, error) {
	telegramUserID, ok := ctx.Get(TelegramUserIDContextKey).(int64)
	if !ok {
		return 0, errors.New("cannot get from echo.Context TelegramUserID")
	}
	return telegramUserID, nil

}

func GetTelegramUserIDFromPathParams(ctx echo.Context) (int64, error) {
	id := ctx.Param(TelegramUserIDPathParamKey)

	return strconv.ParseInt(id, 10, 64)
}

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
		telegramID, err := GetTelegramUserIDFromPathParams(context)
		if err != nil {
			const status = http.StatusBadRequest
			return context.String(status, http.StatusText(status))
		}

		oAuthToken, err := m.userUseCase.GetOrRefreshOAuthAccessTokenByTelegramUserID(telegramID)
		if err != nil {
			switch concreteErr := err.(type) {
			case repository.OAuthError, repository.UserEntityError:
				const status = http.StatusForbidden
				return context.String(status, http.StatusText(status))
			default:
				return errors.WithStack(concreteErr)
			}
		}

		context.Set(OAuthAccessTokenContextKey, oAuthToken)
		context.Set(TelegramUserIDContextKey, telegramID)

		return next(context)
	}
}

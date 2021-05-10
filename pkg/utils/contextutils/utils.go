package contextutils

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const (
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

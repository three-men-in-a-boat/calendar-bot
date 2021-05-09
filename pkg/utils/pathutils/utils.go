package pathutils

import (
	"github.com/labstack/echo/v4"
	"strconv"
)

const (
	TelegramUserIDRouteKey     = ":telegramUserID"
	TelegramUserIDPathParamKey = "telegramUserID"
)

func GetTelegramUserIDFromPathParams(ctx echo.Context) (int64, error) {
	id := ctx.Param(TelegramUserIDPathParamKey)
	return strconv.ParseInt(id, 10, 64)
}

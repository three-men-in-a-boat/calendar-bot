package middlewares

import (
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func LogErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {
		if err := next(context); err != nil {
			zap.S().Errorf("%v", err)
			return err
		}
		return nil
	}
}

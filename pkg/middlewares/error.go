package middlewares

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func LogErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {
		if resErr := next(context); resErr != nil {
			switch err := resErr.(type) {
			case *echo.HTTPError:
				if err.Internal != nil {
					zap.S().Errorf("echo.HTTPError: %+v", err)
				} else {
					zap.S().Debugf("echo.HTTPError: %+v", err)
				}
			default:
				zap.S().Errorf("%+v", err)
			}
			return resErr
		}
		return nil
	}
}

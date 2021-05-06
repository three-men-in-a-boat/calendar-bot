package customerrors

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
)

func HandleCloser(outerErr error, closer io.Closer) error {
	if newErr := closer.Close(); newErr != nil {
		if outerErr != nil {
			outerErr = errors.Wrap(outerErr, newErr.Error())
		} else {
			outerErr = errors.WithStack(newErr)
		}
	}
	return outerErr
}

func HandlerError(err error) {
	zap.S().Error(err)
}

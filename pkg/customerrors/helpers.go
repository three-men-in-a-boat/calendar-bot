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

func HandlerError(err error, chatId *int64, msgId *int) {
	if chatId != nil {
		zap.S().Error("Chat ", chatId)
	}
	if msgId != nil {
		zap.S().Error("Message ", msgId)
	}
	zap.S().Error(err)
}

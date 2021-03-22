package customerrors

import (
	"github.com/pkg/errors"
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

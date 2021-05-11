package oauth

import (
	"fmt"
	"github.com/pkg/errors"
)

type Error struct {
	error
}

var (
	StateKeyDoesNotExist    = Error{errors.New("state does not exist in redis")}
	AccessTokenDoesNotExist = Error{errors.New("OAuth access token does not exist in redis")}
)

type APIResponseErr struct {
	ErrorName        string `json:"error,omitempty"`
	ErrorCode        int    `json:"error_code,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func (o *APIResponseErr) Error() string {
	if o == nil {
		return "<nil>"
	}
	return fmt.Sprintf(
		"oauth.APIResponseErr: error='%s', error_code=%d, error_description='%s'",
		o.ErrorName, o.ErrorCode, o.ErrorDescription,
	)
}

func (o *APIResponseErr) IsError() bool {
	if o == nil || o.ErrorName == "" {
		return false
	}
	return true
}

func (o *APIResponseErr) GetError() *APIResponseErr {
	return o
}

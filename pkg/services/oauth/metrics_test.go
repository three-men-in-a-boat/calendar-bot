package oauth

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricStatusFromErr(t *testing.T) {
	dummyErr := fmt.Errorf("dummy error")
	data := []struct {
		err      error
		expected string
	}{
		{nil, "ok"},
		{AccessTokenDoesNotExist, "access_token_does_not_exist"},
		{StateKeyDoesNotExist, "state_key_does_not_exist"},
		{Error{dummyErr}, "unknown_oauth_service_err"},
		{&APIResponseErr{}, "oauth_api_err_0"},
		{redis.Nil, "redis_nil"},
		{redis.TxFailedErr, "redis_err"},
		{dummyErr, "unknown_err"},
	}

	for _, testCase := range data {
		actual := metricStatusFromErr(testCase.err)
		assert.Equal(t, testCase.expected, actual)
	}
}

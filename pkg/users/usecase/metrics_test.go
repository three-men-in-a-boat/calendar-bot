package usecase

import (
	"fmt"
	"github.com/calendar-bot/pkg/services/oauth"
	"github.com/calendar-bot/pkg/users/repository"
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
		{repository.UserDoesNotExist, "user_does_not_exist"},
		{oauth.AccessTokenDoesNotExist, "access_token_does_not_exist"},
		{oauth.StateKeyDoesNotExist, "state_key_does_not_exist"},
		{&oauth.APIResponseErr{}, "oauth_api_err_0"},
		{dummyErr, "unknown_err"},
	}

	for _, testCase := range data {
		actual := metricStatusFromErr(testCase.err)
		assert.Equal(t, testCase.expected, actual)
	}
}

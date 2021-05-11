package usecase

import (
	"fmt"
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
		{dummyErr, "unknown_err"},
	}

	for _, testCase := range data {
		actual := metricStatusFromErr(testCase.err)
		assert.Equal(t, testCase.expected, actual)
	}
}

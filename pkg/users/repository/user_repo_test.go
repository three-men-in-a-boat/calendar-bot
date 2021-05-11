package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPostgresPlaceholdersForInSQLExpression(t *testing.T) {
	positiveData := []struct {
		count    int
		expected string
	}{
		{1, "$1"},
		{2, "$1,$2"},
		{5, "$1,$2,$3,$4,$5"},
	}

	for _, testCase := range positiveData {
		actual, err := postgresPlaceholdersForInSQLExpression(testCase.count)
		assert.NoError(t, err)
		assert.Equal(t, testCase.expected, actual)
	}

	for _, count := range []int{-2, -1, 0} {
		_, err := postgresPlaceholdersForInSQLExpression(count)
		assert.Error(t, err)
	}
}

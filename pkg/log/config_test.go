package log

import (
	"github.com/bxcodec/faker/v3"
	"github.com/calendar-bot/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type logConfigTestSuite struct {
	suite.Suite
}

func TestInit(t *testing.T) {
	var testSuite logConfigTestSuite
	suite.Run(t, &testSuite)
}

func (s *logConfigTestSuite) testConfigurationLoader(
	expected utils.Enver,
	configLoader func() utils.Enver) {

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual := configLoader()

	assert.Equal(s.T(), expected, actual)
}

func (s *logConfigTestSuite) setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(s.T(), err)
	}
}

func (s *logConfigTestSuite) unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		require.NoError(s.T(), err)
	}
}

func (s *logConfigTestSuite) TestLogConfig() {
	expected := &Config{}
	require.NoError(s.T(), faker.FakeData(expected))

	s.testConfigurationLoader(expected, func() utils.Enver {
		actual := LoadLogConfig()
		return &actual
	})
}

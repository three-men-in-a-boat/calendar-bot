package oauth

import (
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type enver interface {
	ToEnv() map[string]string
}

type oauthConfigTestSuite struct {
	suite.Suite
}

func TestInit(t *testing.T) {
	var testSuite oauthConfigTestSuite
	suite.Run(t, &testSuite)
}

func (s *oauthConfigTestSuite) testConfigurationLoader(
	expected enver,
	configLoader func() enver) {

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual := configLoader()

	assert.Equal(s.T(), expected, actual)
}

func (s *oauthConfigTestSuite) setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(s.T(), err)
	}
}

func (s *oauthConfigTestSuite) unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		require.NoError(s.T(), err)
	}
}

func (s *oauthConfigTestSuite) TestOAuthConfig() {
	expected := &Config{}
	require.NoError(s.T(), faker.FakeData(expected))

	s.testConfigurationLoader(expected, func() enver {
		actual, err := LoadOAuthConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

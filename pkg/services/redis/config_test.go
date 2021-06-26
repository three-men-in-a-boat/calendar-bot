package redis

import (
	"github.com/bxcodec/faker/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"strconv"
	"testing"
)

type enver interface {
	ToEnv() map[string]string
}

type redisConfigTestSuite struct {
	suite.Suite
}

func TestInit(t *testing.T) {
	var testSuite redisConfigTestSuite
	suite.Run(t, &testSuite)
}

func (s *redisConfigTestSuite) testConfigurationLoader(
	expected enver,
	configLoader func() enver) {

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual := configLoader()

	assert.Equal(s.T(), expected, actual)
}

func (s *redisConfigTestSuite) setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(s.T(), err)
	}
}

func (s *redisConfigTestSuite) unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		require.NoError(s.T(), err)
	}
}

func (s *redisConfigTestSuite) TestRedisConfig() {
	expected := Config{}
	require.NoError(s.T(), faker.FakeData(&expected))

	expected = NewConfig(expected.Address, expected.Password, expected.DB)

	s.testConfigurationLoader(&expected, func() enver {
		actual, err := LoadRedisConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *redisConfigTestSuite) TestRedisEmptyDB() {
	expected := Config{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvRedisDB] = ""

	expected = NewConfig(expected.Address, expected.Password, 0)

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual, err := LoadRedisConfig()
	require.NoError(s.T(), err)

	assert.Equal(s.T(), expected, actual)
}

func (s *redisConfigTestSuite) TestRedisInvalidDB() {
	expected := Config{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvRedisDB] = "invalid"

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	_, err := LoadRedisConfig()
	convertErr := errors.Cause(err).(*strconv.NumError)

	assert.Error(s.T(), convertErr)
}

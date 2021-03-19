package config

import (
	"github.com/bxcodec/faker/v3"
	//github.com/brianvoe/gofakeit/v6
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"strconv"
	"testing"
)

type configTestSuite struct {
	suite.Suite
}

func TestInit(t *testing.T) {
	var testSuite configTestSuite
	suite.Run(t, &testSuite)
}

func (s *configTestSuite) TestAppConfigAppEnvironmentValueRandom() {
	expected := App{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual, err := LoadAppConfig()
	require.NoError(s.T(), err)

	expected.Environment = AppEnvironmentProd
	require.Equal(s.T(), expected, actual)
}

func (s *configTestSuite) TestAppConfigAppEnvironmentValueProd() {
	expected := &App{}
	require.NoError(s.T(), faker.FakeData(expected))

	expected.Environment = AppEnvironmentProd

	s.testConfigurationLoader(expected, func() Enver {
		actual, err := LoadAppConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *configTestSuite) TestAppConfigAppEnvironmentValueDev() {
	expected := &App{}
	require.NoError(s.T(), faker.FakeData(expected))

	expected.Environment = AppEnvironmentDev

	s.testConfigurationLoader(expected, func() Enver {
		actual, err := LoadAppConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *configTestSuite) TestDBConfig() {
	expected := &DB{}
	require.NoError(s.T(), faker.FakeData(expected))

	s.testConfigurationLoader(expected, func() Enver {
		actual, err := LoadDBConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *configTestSuite) TestDBConfigEmptyMaxOpenConnections() {
	expected := DB{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvDBMaxOpenConnections] = ""
	expected.MaxOpenConnections = 10

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual, err := LoadDBConfig()
	require.NoError(s.T(), err)

	require.Equal(s.T(), expected, actual)
}

func (s *configTestSuite) TestDBConfigInvalidMaxOpenConnections() {
	expected := DB{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvDBMaxOpenConnections] = "invalid"

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	_, err := LoadDBConfig()
	convertErr := errors.Cause(err).(*strconv.NumError)

	require.Error(s.T(), convertErr)
}

func (s *configTestSuite) TestLogConfig() {
	expected := &Log{}
	require.NoError(s.T(), faker.FakeData(expected))

	s.testConfigurationLoader(expected, func() Enver {
		actual := LoadLogConfig()
		return &actual
	})
}

func (s *configTestSuite) TestOAuthConfig() {
	expected := &OAuth{}
	require.NoError(s.T(), faker.FakeData(expected))

	s.testConfigurationLoader(expected, func() Enver {
		actual := LoadOAuthConfig()
		return &actual
	})
}

func (s *configTestSuite) TestRedisConfig() {
	expected := &Redis{}
	require.NoError(s.T(), faker.FakeData(expected))

	s.testConfigurationLoader(expected, func() Enver {
		actual, err := LoadRedisConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *configTestSuite) TestRedisEmptyDB() {
	expected := Redis{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvRedisDB] = ""
	expected.DB = 0

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual, err := LoadRedisConfig()
	require.NoError(s.T(), err)

	require.Equal(s.T(), expected, actual)
}

func (s *configTestSuite) TestRedisInvalidDB() {
	expected := Redis{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvRedisDB] = "invalid"

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	_, err := LoadRedisConfig()
	convertErr := errors.Cause(err).(*strconv.NumError)

	require.Error(s.T(), convertErr)
}

func (s *configTestSuite) testConfigurationLoader(
	expected Enver,
	configLoader func() Enver) {

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual := configLoader()

	require.Equal(s.T(), expected, actual)
}

func (s *configTestSuite) setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(s.T(), err)
	}
}

func (s *configTestSuite) unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		require.NoError(s.T(), err)
	}
}

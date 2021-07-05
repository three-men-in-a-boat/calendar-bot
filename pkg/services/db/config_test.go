package db

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

type dbConfigTestSuite struct {
	suite.Suite
}

func TestInit(t *testing.T) {
	var testSuite dbConfigTestSuite
	suite.Run(t, &testSuite)
}

func (s *dbConfigTestSuite) testConfigurationLoader(
	expected enver,
	configLoader func() enver) {

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual := configLoader()

	assert.Equal(s.T(), expected, actual)
}

func (s *dbConfigTestSuite) setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(s.T(), err)
	}
}

func (s *dbConfigTestSuite) unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		require.NoError(s.T(), err)
	}
}

func (s *dbConfigTestSuite) TestDBConfig() {
	expected := &Config{}
	require.NoError(s.T(), faker.FakeData(expected))

	s.testConfigurationLoader(expected, func() enver {
		actual, err := LoadDBConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *dbConfigTestSuite) TestDBConfigEmptyMaxOpenConnections() {
	expected := Config{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvDBMaxOpenConnections] = ""
	expected.MaxOpenConnections = 10

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual, err := LoadDBConfig()
	require.NoError(s.T(), err)

	assert.Equal(s.T(), expected, actual)
}

func (s *dbConfigTestSuite) TestDBConfigInvalidMaxOpenConnections() {
	expected := Config{}
	require.NoError(s.T(), faker.FakeData(&expected))

	envs := expected.ToEnv()
	envs[EnvDBMaxOpenConnections] = "invalid"

	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	_, err := LoadDBConfig()
	convertErr := errors.Cause(err).(*strconv.NumError)

	assert.Error(s.T(), convertErr)
}

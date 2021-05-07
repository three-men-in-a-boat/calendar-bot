package config

import (
	"github.com/bxcodec/faker/v3"
	"github.com/calendar-bot/pkg/services/redis"
	"github.com/calendar-bot/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type appConfigTestSuite struct {
	suite.Suite
}

func TestInit(t *testing.T) {
	var testSuite appConfigTestSuite
	suite.Run(t, &testSuite)
}

func (s *appConfigTestSuite) testConfigurationLoader(
	expected utils.Enver,
	configLoader func() utils.Enver) {

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual := configLoader()

	assert.Equal(s.T(), expected, actual)
}

func (s *appConfigTestSuite) setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(s.T(), err)
	}
}

func (s *appConfigTestSuite) unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		require.NoError(s.T(), err)
	}
}

func (s *appConfigTestSuite) generateFakeAppConfig() AppConfig {
	config := AppConfig{}
	require.NoError(s.T(), faker.FakeData(&config))

	config.BotRedis = redis.NewBotConfig(
		config.Redis.Address,
		config.Redis.Password,
		config.BotRedis.DB,
	)
	config.Redis = redis.NewConfig(
		config.Redis.Address,
		config.Redis.Password,
		config.Redis.DB,
	)
	return config
}

func (s *appConfigTestSuite) TestAppConfigAppEnvironmentValueRandom() {
	expected := s.generateFakeAppConfig()

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual, err := LoadAppConfig()
	assert.NoError(s.T(), err)

	expected.Environment = AppEnvironmentProd
	assert.Equal(s.T(), expected, actual)
}

func (s *appConfigTestSuite) TestAppConfigAppEnvironmentValueProd() {
	expected := s.generateFakeAppConfig()

	expected.Environment = AppEnvironmentProd

	s.testConfigurationLoader(&expected, func() utils.Enver {
		actual, err := LoadAppConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *appConfigTestSuite) TestAppConfigAppEnvironmentValueDev() {
	expected := s.generateFakeAppConfig()

	expected.Environment = AppEnvironmentDev

	s.testConfigurationLoader(&expected, func() utils.Enver {
		actual, err := LoadAppConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

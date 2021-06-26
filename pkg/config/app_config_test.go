package config

import (
	"github.com/bxcodec/faker/v3"
	"github.com/calendar-bot/pkg/services/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type enver interface {
	ToEnv() map[string]string
}

type appConfigTestSuite struct {
	suite.Suite
}

func TestInit(t *testing.T) {
	var testSuite appConfigTestSuite
	suite.Run(t, &testSuite)
}

func (s *appConfigTestSuite) testConfigurationLoader(
	expected enver,
	configLoader func() enver) {

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

	config.BotDefaultUserTimezone = defaultBotUserTimezoneValue
	config.OAuth.LinkExpireIn = 15 * time.Minute

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

func (s *appConfigTestSuite) TestAppConfigInvalidTimezone() {
	expected := s.generateFakeAppConfig()
	expected.BotDefaultUserTimezone = "invalid-timezone"

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	_, err := LoadAppConfig()
	assert.Error(s.T(), err)
}

func (s *appConfigTestSuite) TestAppConfigEmptyTimezone() {
	expected := s.generateFakeAppConfig()
	expected.BotDefaultUserTimezone = ""

	envs := expected.ToEnv()
	s.setEnvs(envs)
	defer s.unsetEnvs(envs)

	actual, err := LoadAppConfig()
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), actual.BotDefaultUserTimezone, defaultBotUserTimezoneValue)
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

	s.testConfigurationLoader(&expected, func() enver {
		actual, err := LoadAppConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

func (s *appConfigTestSuite) TestAppConfigAppEnvironmentValueDev() {
	expected := s.generateFakeAppConfig()

	expected.Environment = AppEnvironmentDev

	s.testConfigurationLoader(&expected, func() enver {
		actual, err := LoadAppConfig()
		require.NoError(s.T(), err)
		return &actual
	})
}

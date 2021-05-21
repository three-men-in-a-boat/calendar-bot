package config

import (
	"github.com/calendar-bot/pkg/log"
	"github.com/calendar-bot/pkg/services/db"
	"github.com/calendar-bot/pkg/services/oauth"
	"github.com/calendar-bot/pkg/services/redis"
	"github.com/pkg/errors"
	"os"
	"time"
)

const (
	EnvAppEnvironment = "APP_ENVIRONMENT"
	EnvAppAddress     = "APP_ADDRESS"
	EnvParseAddress   = "PARSER_BACKEND_URL"
)

const (
	EnvBotAddress             = "BOT_ADDRESS"
	EnvBotToken               = "BOT_TOKEN"
	EnvBotWebhookUrl          = "BOT_WEBHOOK_URL"
	EnvBotDefaultUserTimezone = "BOT_DEFAULT_USER_TIMEZONE"
)

const (
	AppEnvironmentDev  string = "dev"
	AppEnvironmentProd string = "prod"
)

const (
	defaultBotUserTimezoneValue = "Europe/Moscow"
)

type AppConfig struct {
	Address                string `validate:"optional,dialstring"`
	Environment            string `validate:"optional,in(dev|prod)"`
	BotAddress             string `validate:"optional"`
	BotToken               string
	BotWebhookUrl          string
	BotDefaultUserTimezone string
	DB                     db.Config
	ParseAddress           string
	Redis                  redis.Config
	BotRedis               redis.Config
	OAuth                  oauth.Config
	Log                    log.Config
}

func LoadAppConfig() (AppConfig, error) {
	address := os.Getenv(EnvAppAddress)
	environment := os.Getenv(EnvAppEnvironment)
	parseAddress := os.Getenv(EnvParseAddress)

	botAddress := os.Getenv(EnvBotAddress)
	botToken := os.Getenv(EnvBotToken)
	botWebhookUrl := os.Getenv(EnvBotWebhookUrl)
	botDefaultUserTimezone := os.Getenv(EnvBotDefaultUserTimezone)

	if address == "" {
		address = ":8080"
	}

	if botAddress == "" {
		address = ":2000"
	}
	if botDefaultUserTimezone == "" {
		botDefaultUserTimezone = defaultBotUserTimezoneValue
	}

	if _, err := time.LoadLocation(botDefaultUserTimezone); err != nil {
		return AppConfig{},
			errors.WithMessagef(
				err, "failed to load timezone %q from local timezones DB", botDefaultUserTimezone)
	}

	switch environment {
	case AppEnvironmentProd, AppEnvironmentDev:
		// nickeskov: app environment ok
	default:
		environment = AppEnvironmentProd
	}

	dbConfig, err := db.LoadDBConfig()
	if err != nil {
		return AppConfig{}, errors.WithMessage(err, "failed to load db config")
	}

	redisConfig, err := redis.LoadRedisConfig()
	if err != nil {
		return AppConfig{}, errors.WithMessage(err, "failed to load redis config")
	}

	botRedisConfig, err := redis.LoadBotRedisConfig()
	if err != nil {
		return AppConfig{}, errors.WithMessage(err, "failed to load bot redis config")
	}

	oauthConfig, err := oauth.LoadOAuthConfig()
	if err != nil {
		return AppConfig{}, errors.WithMessage(err, "failed to load oauth config")
	}

	// TODO(nickeskov): validate struct

	return AppConfig{
		Address:                address,
		BotAddress:             botAddress,
		BotToken:               botToken,
		BotWebhookUrl:          botWebhookUrl,
		BotDefaultUserTimezone: botDefaultUserTimezone,
		Environment:            environment,
		DB:                     dbConfig,
		ParseAddress:           parseAddress,
		Redis:                  redisConfig,
		BotRedis:               botRedisConfig,
		OAuth:                  oauthConfig,
		Log:                    log.LoadLogConfig(),
	}, nil
}

func (app *AppConfig) ToEnv() map[string]string {
	envMaps := [...]map[string]string{
		app.DB.ToEnv(),
		app.BotRedis.ToEnv(),
		app.Redis.ToEnv(),
		app.OAuth.ToEnv(),
		app.Log.ToEnv(),
	}

	capacity := 0
	for _, m := range envMaps {
		capacity += len(m)
	}

	ret := make(map[string]string, capacity)

	for _, m := range envMaps {
		for key, value := range m {
			ret[key] = value
		}
	}

	ret[EnvAppAddress] = app.Address
	ret[EnvAppEnvironment] = app.Environment
	ret[EnvParseAddress] = app.ParseAddress

	ret[EnvBotAddress] = app.BotAddress
	ret[EnvBotToken] = app.BotToken
	ret[EnvBotWebhookUrl] = app.BotWebhookUrl
	ret[EnvBotDefaultUserTimezone] = app.BotDefaultUserTimezone

	return ret
}

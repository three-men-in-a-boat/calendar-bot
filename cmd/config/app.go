package config

import (
	"github.com/pkg/errors"
	"os"
)

const (
	EnvAppEnvironment = "APP_ENVIRONMENT"
	EnvAppAddress     = "APP_ADDRESS"
)

const (
	EnvBotAddress    = "BOT_ADDRESS"
	EnvBotToken      = "BOT_TOKEN"
	EnvBotWebhookUrl = "BOT_WEBHOOK_URL"
)

const (
	AppEnvironmentDev  string = "dev"
	AppEnvironmentProd string = "prod"
)

type App struct {
	Address       string `validate:"optional,dialstring"`
	Environment   string `validate:"optional,in(dev|prod)"`
	BotAddress    string `validate:"optional"`
	BotToken      string
	BotWebhookUrl string
	DB            DB
	Redis         Redis
	BotRedis      Redis
	OAuth         OAuth
	Log           Log
}

func LoadAppConfig() (App, error) {
	address := os.Getenv(EnvAppAddress)
	environment := os.Getenv(EnvAppEnvironment)

	botAddress := os.Getenv(EnvBotAddress)
	botToken := os.Getenv(EnvBotToken)
	botWebhookUrl := os.Getenv(EnvBotWebhookUrl)

	if address == "" {
		address = ":8080"
	}

	if botAddress == "" {
		address = ":2000"
	}

	switch environment {
	case AppEnvironmentProd, AppEnvironmentDev:
		// nickeskov: app environment ok
	default:
		environment = AppEnvironmentProd
	}

	db, err := LoadDBConfig()
	if err != nil {
		return App{}, errors.WithMessage(err, "failed to load DB config")
	}

	redis, err := LoadRedisConfig()
	if err != nil {
		return App{}, errors.WithMessage(err, "failed to load Redis config")
	}

	botRedis, err := LoadBotRedisConfig()
	if err != nil {
		return App{}, errors.WithMessage(err, "failed to load bot redis config")
	}
	// TODO(nickeskov): validate struct

	return App{
		Address:       address,
		BotAddress:    botAddress,
		BotToken:      botToken,
		BotWebhookUrl: botWebhookUrl,
		Environment:   environment,
		DB:            db,
		Redis:         redis,
		BotRedis:      botRedis,
		OAuth:         LoadOAuthConfig(),
		Log:           LoadLogConfig(),
	}, nil
}

func (app *App) ToEnv() map[string]string {
	envMaps := [...]map[string]string{
		app.DB.ToEnv(),
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

	ret[EnvBotAddress] = app.BotAddress
	ret[EnvBotToken] = app.BotToken
	ret[EnvBotWebhookUrl] = app.BotWebhookUrl

	return ret
}

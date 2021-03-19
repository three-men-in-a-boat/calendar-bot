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
	AppEnvironmentDev  string = "dev"
	AppEnvironmentProd string = "prod"
)

type App struct {
	Address     string `validate:"optional,dialstring"`
	Environment string `validate:"optional,in(dev|prod)"`
	DB          DB
	Redis       Redis
	OAuth       OAuth
	Log         Log
}

func LoadAppConfig() (App, error) {
	address := os.Getenv(EnvAppAddress)
	environment := os.Getenv(EnvAppEnvironment)

	if address == "" {
		address = ":8080"
	}

	switch environment {
	case AppEnvironmentProd, AppEnvironmentDev:
		// nickeskov: app environment ok
	default:
		environment = AppEnvironmentProd
	}

	db, err := LoadDBConfig()
	if err != nil {
		return App{}, errors.WithMessage(err, "cannot load DB config")
	}

	redis, err := LoadRedisConfig()
	if err != nil {
		return App{}, errors.WithMessage(err, "cannot load Redis config")
	}

	// TODO(nickeskov): validate struct

	return App{
		Address:     address,
		Environment: environment,
		DB:          db,
		Redis:       redis,
		OAuth:       LoadOAuthConfig(),
		Log:         LoadLogConfig(),
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

	return ret
}

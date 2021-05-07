package redis

import (
	"github.com/pkg/errors"
	"os"
	"strconv"
)

const (
	EnvRedisAddress  = "REDIS_ADDRESS"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvRedisDB       = "REDIS_DB"
	EnvBotRedisDB    = "REDIS_BOT_DB"
)

const (
	appRedisDB = iota + 1
	botRedisDB
)

type Config struct {
	Address     string `valid:"dialstring"`
	Password    string `valid:"-"`
	DB          int    `valid:"-"`
	redisDBType int    `valid:"-"`
}

func NewConfig(address, password string, db int) Config {
	return Config{
		Address:     address,
		Password:    password,
		DB:          db,
		redisDBType: appRedisDB,
	}
}
func LoadRedisConfig() (Config, error) {
	address := os.Getenv(EnvRedisAddress)
	password := os.Getenv(EnvRedisPassword)

	db := 0
	if dbStr := os.Getenv(EnvRedisDB); dbStr != "" {
		num, err := strconv.Atoi(dbStr)
		if err != nil {
			return Config{}, errors.WithMessagef(err, "failed to parse %s environment variable as int", EnvRedisDB)
		}
		db = num
	}

	// TODO(nickeskov): validate struct

	return NewConfig(address, password, db), nil
}

func NewBotConfig(address, password string, db int) Config {
	return Config{
		Address:     address,
		Password:    password,
		DB:          db,
		redisDBType: botRedisDB,
	}
}

func LoadBotRedisConfig() (Config, error) {
	address := os.Getenv(EnvRedisAddress)
	password := os.Getenv(EnvRedisPassword)

	db := 0
	if dbStr := os.Getenv(EnvBotRedisDB); dbStr != "" {
		num, err := strconv.Atoi(dbStr)
		if err != nil {
			return Config{},
				errors.WithMessagef(err, "failed to parse %s environment variable as int", EnvRedisDB)
		}
		db = num
	}

	return NewBotConfig(address, password, db), nil
}

func (r *Config) ToEnv() map[string]string {
	envs := map[string]string{
		EnvRedisAddress:  r.Address,
		EnvRedisPassword: r.Password,
	}

	var redisDBEnv string
	switch r.redisDBType {
	case appRedisDB:
		redisDBEnv = EnvRedisDB
	case botRedisDB:
		redisDBEnv = EnvBotRedisDB
	default:
		return envs
	}

	envs[redisDBEnv] = strconv.Itoa(r.DB)

	return envs
}

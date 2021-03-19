package config

import (
	"github.com/pkg/errors"
	"os"
	"strconv"
)

//"github.com/go-redis/redis/v8"

const (
	EnvRedisAddress  = "REDIS_ADDRESS"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvRedisDB       = "REDIS_DB"
)

type Redis struct {
	Address  string `valid:"dialstring"`
	Password string `valid:"-"`
	DB       int    `valid:"-"`
}

func LoadRedisConfig() (Redis, error) {
	address := os.Getenv(EnvRedisAddress)
	password := os.Getenv(EnvRedisPassword)

	db := 0
	if dbStr := os.Getenv(EnvRedisDB); dbStr != "" {
		num, err := strconv.Atoi(dbStr)
		if err != nil {
			return Redis{}, errors.WithMessagef(err, "cannot parse %s environment variable as int", EnvRedisDB)
		}
		db = num
	}

	// TODO(nickeskov): validate struct

	return Redis{
		Address:  address,
		Password: password,
		DB:       db,
	}, nil
}

func (r *Redis) ToEnv() map[string]string {
	return map[string]string{
		EnvRedisAddress:  r.Address,
		EnvRedisPassword: r.Password,
		EnvRedisDB:       strconv.Itoa(r.DB),
	}
}

package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

func ConnectToRedis(conf *Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Address,
		Password: conf.Password,
		DB:       conf.DB,
	})
	if err := client.Ping(context.TODO()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

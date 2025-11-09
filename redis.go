package bint

import (
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

func initRedis(cfg RedisConfig) (r *redis.Client, err error) {
	r = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return r, r.Ping(ctx).Err()
}

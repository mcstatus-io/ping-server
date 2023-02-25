package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Client *redis.Client
}

func (r *Redis) Connect() error {
	r.Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Username: config.Redis.Username,
		Password: config.Redis.Password,
		DB:       config.Redis.Database,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Client.Ping(ctx).Err()
}

func (r *Redis) Exists(key string) (bool, error) {
	if r.Client == nil {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	res := r.Client.Exists(ctx, key)

	if err := res.Err(); err != nil {
		return false, err
	}

	val, err := res.Result()

	return val == 1, err
}

func (r *Redis) TTL(key string) (time.Duration, error) {
	if r.Client == nil {
		return 0, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	res := r.Client.TTL(ctx, key)

	if err := res.Err(); err != nil {
		return 0, err
	}

	return res.Result()
}

func (r *Redis) GetString(key string) (string, error) {
	if r.Client == nil {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	res := r.Client.Get(ctx, key)

	if err := res.Err(); err != nil {
		return "", nil
	}

	return res.Result()
}

func (r *Redis) GetBytes(key string) ([]byte, error) {
	if r.Client == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	res := r.Client.Get(ctx, key)

	if err := res.Err(); err != nil {
		return nil, err
	}

	return res.Bytes()
}

func (r *Redis) Set(key string, value interface{}, ttl time.Duration) error {
	if r.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Increment(key string) error {
	if r.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Client.Incr(ctx, key).Err()
}

func (r *Redis) Close() error {
	if r.Client == nil {
		return nil
	}

	return r.Client.Close()
}

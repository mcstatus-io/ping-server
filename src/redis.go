package main

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Client *redis.Client
}

func (r *Redis) Connect() error {
	opts, err := redis.ParseURL(*config.Redis)

	if err != nil {
		return err
	}

	r.Client = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Client.Ping(ctx).Err()
}

func (r *Redis) Get(key string) ([]byte, time.Duration, error) {
	if r.Client == nil {
		return nil, 0, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	p := r.Client.Pipeline()

	value := p.Get(ctx, key)
	ttl := p.TTL(ctx, key)

	if _, err := p.Exec(ctx); err != nil {
		if err == redis.Nil {
			return nil, 0, nil
		}

		return nil, 0, err
	}

	data, err := value.Bytes()

	return data, ttl.Val(), err
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

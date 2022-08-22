package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Enabled bool
	Client  *redis.Client
}

func (r *Redis) SetEnabled(value bool) {
	r.Enabled = value
}

func (r *Redis) Connect(uri string) error {
	if !r.Enabled {
		return nil
	}

	opts, err := redis.ParseURL(uri)

	if err != nil {
		return err
	}

	r.Client = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Client.Ping(ctx).Err()
}

func (r *Redis) Exists(key string) (bool, error) {
	if !r.Enabled {
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
	if !r.Enabled {
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
	if !r.Enabled {
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
	if !r.Enabled {
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
	if !r.Enabled {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) SetJSON(key string, value interface{}, ttl time.Duration) error {
	if !r.Enabled {
		return nil
	}

	data, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return r.Set(key, data, ttl)
}

func (r *Redis) Close() error {
	if !r.Enabled {
		return nil
	}

	return r.Client.Close()
}

package main

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Conn *redis.Client
}

func (r *Redis) Connect(uri string) error {
	opts, err := redis.ParseURL(uri)

	if err != nil {
		return err
	}

	conn := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	if err = conn.Ping(ctx).Err(); err != nil {
		return err
	}

	r.Conn = conn

	return nil
}

func (r *Redis) TTL(key string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	val, err := r.Conn.TTL(ctx, key).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}

		return 0, err
	}

	return val, nil
}

func (r *Redis) Exists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	val, err := r.Conn.Exists(ctx, key).Result()

	return val == 1, err
}

func (r *Redis) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	val, err := r.Conn.Get(ctx, key).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}

		return "", err
	}

	return val, nil
}

func (r *Redis) GetBytes(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	result := r.Conn.Get(ctx, key)

	if err := result.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}

		return nil, err
	}

	return result.Bytes()
}

func (r *Redis) Set(key string, value interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Conn.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) GetValueAndTTL(key string) (bool, string, time.Duration, error) {
	exists, err := r.Exists(key)

	if err != nil {
		return false, "", 0, err
	}

	if !exists {
		return false, "", 0, nil
	}

	value, err := r.Get(key)

	if err != nil {
		return false, "", 0, err
	}

	ttl, err := r.TTL(key)

	return true, value, ttl, err
}

func (r *Redis) Keys(pattern string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	res := r.Conn.Keys(ctx, pattern)

	if err := res.Err(); err != nil {
		return nil, err
	}

	return res.Result()
}

func (r *Redis) Delete(keys ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return r.Conn.Del(ctx, keys...).Err()
}

func (r *Redis) Close() error {
	return r.Conn.Close()
}

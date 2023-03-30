package main

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

const defaultTimeout = 5 * time.Second

// Redis is a wrapper around the Redis client.
type Redis struct {
	Client *redis.Client
}

// Connect establishes a connection to the Redis server using the configuration.
func (r *Redis) Connect() error {
	if config.Redis == nil {
		return errors.New("missing Redis configuration")
	}

	opts, err := redis.ParseURL(*config.Redis)
	if err != nil {
		return err
	}

	r.Client = redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return r.Client.Ping(ctx).Err()
}

// Get retrieves the value and TTL for a given key.
func (r *Redis) Get(key string) ([]byte, time.Duration, error) {
	if r.Client == nil {
		return nil, 0, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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

// Set sets the value and TTL for a given key.
func (r *Redis) Set(key string, value interface{}, ttl time.Duration) error {
	if r.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return r.Client.Set(ctx, key, value, ttl).Err()
}

// Increment increments the integer value of a key by 1.
func (r *Redis) Increment(key string) error {
	if r.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return r.Client.Incr(ctx, key).Err()
}

// Close closes the Redis client connection.
func (r *Redis) Close() error {
	if r.Client == nil {
		return nil
	}

	return r.Client.Close()
}

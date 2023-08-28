package main

import (
	"context"
	"errors"
	"time"

	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	redsyncredislib "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

const defaultTimeout = 5 * time.Second

// Redis is a wrapper around the Redis client.
type Redis struct {
	Client     *redis.Client
	Pool       *redsyncredis.Pool
	SyncClient *redsync.Redsync
}

// Connect establishes a connection to the Redis server using the configuration.
func (r *Redis) Connect() error {
	if config.Redis == nil {
		return errors.New("missing Redis configuration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	defer cancel()

	opts, err := redis.ParseURL(*config.Redis)

	if err != nil {
		return err
	}

	r.Client = redis.NewClient(opts)

	if err = r.Client.Ping(ctx).Err(); err != nil {
		return err
	}

	pool := redsyncredislib.NewPool(r.Client)

	r.SyncClient = redsync.New(pool)

	return nil
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

// NewMutex creates a new mutually exclusive lock that only one process can hold.
func (r *Redis) NewMutex(name string) *Mutex {
	if r.Client == nil || r.SyncClient == nil {
		return &Mutex{
			Mutex: nil,
		}
	}

	return &Mutex{
		Mutex: r.SyncClient.NewMutex(name),
	}
}

// Close closes the Redis client connection.
func (r *Redis) Close() error {
	if r.Client == nil {
		return nil
	}

	return r.Client.Close()
}

// Mutex is a mutually exclusive lock held across all processes.
type Mutex struct {
	Mutex *redsync.Mutex
}

// Lock will lock the mutex so no other process can hold it.
func (m *Mutex) Lock() error {
	if m.Mutex == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return m.Mutex.LockContext(ctx)
}

// Unlock will allow any other process to obtain a lock with the same key.
func (m *Mutex) Unlock() error {
	if m.Mutex == nil {
		return nil
	}

	_, err := m.Mutex.Unlock()

	return err
}

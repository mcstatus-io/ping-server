package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisConnect(t *testing.T) {
	r := &Redis{}
	config = &Config{
		Redis: nil,
	}

	// Test missing Redis configuration
	err := r.Connect()
	assert.Equal(t, errors.New("missing Redis configuration"), err)

	// Test successful connection
	db, mock := redismock.NewClientMock()
	mock.On("Ping", context.Background()).Return(nil)
	r.Client = db
	err = r.Connect()
	assert.NoError(t, err)
}

func TestRedisGet(t *testing.T) {
	r := &Redis{}
	db, mock := redismock.NewClientMock()
	r.Client = db

	// Test getting value and TTL
	mock.On("Get", context.Background(), "testKey").Return("testValue")
	mock.On("TTL", context.Background(), "testKey").Return(time.Minute, nil)
	value, ttl, err := r.Get("testKey")
	assert.Equal(t, []byte("testValue"), value)
	assert.Equal(t, time.Minute, ttl)
	assert.NoError(t, err)

	// Test key not found
	mock.On("Get", context.Background(), "nonExistentKey").Return(nil)
	mock.On("TTL", context.Background(), "nonExistentKey").Return(time.Duration(0), nil)
	value, ttl, err = r.Get("nonExistentKey")
	assert.Nil(t, value)
	assert.Equal(t, time.Duration(0), ttl)
	assert.NoError(t, err)
}

func TestRedisSet(t *testing.T) {
	r := &Redis{}
	db, mock := redismock.NewClientMock()
	r.Client = db

	// Test setting value and TTL
	mock.On("Set", context.Background(), "testKey", "testValue", time.Minute).Return(nil)
	err := r.Set("testKey", "testValue", time.Minute)
	assert.NoError(t, err)
}

func TestRedisIncrement(t *testing.T) {
	r := &Redis{}
	db, mock := redismock.NewClientMock()
	r.Client = db

	// Test incrementing a key
	mock.On("Incr", context.Background(), "testKey").Return(nil)
	err := r.Increment("testKey")
	assert.NoError(t, err)
}


// Package sessionbolt provides a Redis store for sessions.
//
// This package is only compatible with https://github.com/go-redis/redis.
// But since this package has several different versions for different versions of Redis,
// we do not import any of them,
// but ensure compatibility with any of them through generics.
//
// # How to use?
//
//  1. go get and import relavant go-redis module
//  2. go get and import this module
//  3. define store
//
// Example:
//  import (
//    "github.com/go-redis/redis/v9"
//    // or "github.com/go-redis/redis/v8"
//    "github.com/mr-linch/go-tg/tgb/session/sessionredis"
//  )
//
//  func run(ctx context.Context) error {
// 	  opts, err := redis.ParseURL("redis://localhost:6379")
//    if err != nil {
//      return err
//    }
//    client := redis.NewClient(opts)
//
//    store := sessionredis.NewStore[*redis.StatusCmd, *redis.StringCmd, *redis.IntCmd](client)
//    // use store :)
//  }
package sessionredis

import (
	"context"
	"strings"
	"time"
)

type redisStatusCmd interface {
	Err() error
}

type redisIntCmd interface {
	Err() error
}

type redisStringCmd interface {
	Bytes() ([]byte, error)
}

type Redis[
	SR redisStatusCmd,
	GR redisStringCmd,
	DR redisIntCmd,
] interface {
	Set(ctx context.Context, key string, value interface{}, exp time.Duration) SR
	Get(ctx context.Context, key string) GR
	Del(ctx context.Context, keys ...string) DR
}

type Store[
	SR redisStatusCmd,
	GR redisStringCmd,
	DR redisIntCmd,
] struct {
	rdb Redis[SR, GR, DR]
}

func NewStore[
	SR redisStatusCmd,
	GR redisStringCmd,
	DR redisIntCmd,
](redis Redis[SR, GR, DR]) *Store[SR, GR, DR] {
	return &Store[SR, GR, DR]{
		rdb: redis,
	}
}

func (s *Store[SR, GR, DR]) Set(ctx context.Context, key string, value []byte) error {
	return s.rdb.Set(ctx, key, value, 0).Err()
}

func (s *Store[SR, GR, DR]) Get(ctx context.Context, key string) ([]byte, error) {
	v, err := s.rdb.Get(ctx, key).Bytes()

	// it's ugly, but we should do it for not bind to redis version
	if err != nil && strings.Contains(err.Error(), "redis: nil") {
		return nil, nil
	}

	return v, nil
}

func (s *Store[SR, GR, DR]) Del(ctx context.Context, key string) error {
	return s.rdb.Del(ctx, key).Err()
}

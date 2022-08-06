package sessionredis

import (
	"context"
	"strings"
	"time"
)

type RedisStatusCmd interface {
	Err() error
}

type RedisIntCmd interface {
	Err() error
}

type RedisStringCmd interface {
	Bytes() ([]byte, error)
}

type Redis[
	SR RedisStatusCmd,
	GR RedisStringCmd,
	DR RedisIntCmd,
] interface {
	Set(ctx context.Context, key string, value interface{}, exp time.Duration) SR
	Get(ctx context.Context, key string) GR
	Del(ctx context.Context, keys ...string) DR
}

type Store[
	SR RedisStatusCmd,
	GR RedisStringCmd,
	DR RedisIntCmd,
] struct {
	rdb    Redis[SR, GR, DR]
	prefix string
}

func NewStore[
	SR RedisStatusCmd,
	GR RedisStringCmd,
	DR RedisIntCmd,
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

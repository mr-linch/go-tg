package sessionbolt

import (
	"context"
	"fmt"
	"os"

	"go.etcd.io/bbolt"
)

type Store struct {
	db *bbolt.DB
}

func Open(path string, mode os.FileMode, options *bbolt.Options) (*Store, error) {
	db, err := bbolt.Open(path, mode, options)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	return New(db), nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func New(db *bbolt.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("sessions"))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		return bucket.Put([]byte(key), value)
	})
}

func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte

	if err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("sessions"))
		if bucket == nil {
			return nil
		}

		value = bucket.Get([]byte(key))
		return nil
	}); err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	return value, nil
}

func (s *Store) Del(ctx context.Context, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("sessions"))
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(key))
	})
}

// Package sessionbolt provides a BoltDB store for sessions.
package sessionbolt

import (
	"context"
	"fmt"
	"os"

	"go.etcd.io/bbolt"
)

// Store is a session store that uses BoltDB as backend.
type Store struct {
	db     *bbolt.DB
	bucket []byte
}

// Option of Store.
type Option func(*Store)

// WithBucket sets the bucket name for the store.
// Be default, the bucket name is "sessions".
func WithBucket(bucket string) Option {
	return func(s *Store) {
		s.bucket = []byte(bucket)
	}
}

// Open opens or create and open a BoltDB file.
// See [bbolt.Open] for more details.
func Open(path string, mode os.FileMode, options *bbolt.Options, opts ...Option) (*Store, error) {
	db, err := bbolt.Open(path, mode, options)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	return New(db, opts...), nil
}

// Close closes boltdb store.
func (s *Store) Close() error {
	return s.db.Close()
}

// New creates a new BoltDB store with alredy opened BoltDB.
func New(db *bbolt.DB, opts ...Option) *Store {
	store := &Store{
		db:     db,
		bucket: []byte("sessions"),
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

// Set sets a session value.
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		return bucket.Put([]byte(key), value)
	})
}

// Get gets a session value.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte

	if err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(s.bucket)
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

// Del deletes a session value.
func (s *Store) Del(ctx context.Context, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(s.bucket)
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(key))
	})
}

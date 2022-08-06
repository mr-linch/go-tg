// Package sessionbolt provides a BoltDB store for sessions.
//
// It uses [go.etcd.io/bbolt] package as a backend.
//
// Example:
//
//  db, err := bbolt.Open("db.boltdb", 0666, nil)
//  if err != nil {
//    return err
//  }
//  defer db.Close()
//
//  store := New(db, "sessions")
package sessionbolt

import (
	"context"
	"fmt"

	"go.etcd.io/bbolt"
)

// Store is a session store that uses BoltDB as backend.
type Store struct {
	db     *bbolt.DB
	bucket []byte
}

// New creates a new BoltDB store with alredy opened BoltDB.
func New(db *bbolt.DB, bucket string) *Store {
	return &Store{
		db:     db,
		bucket: []byte(bucket),
	}
}

// Set saves a session in DB
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(s.bucket)
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}

		return bucket.Put([]byte(key), value)
	})
}

// Get get a session from DB by key.
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

// Del deletes a session in DB.
func (s *Store) Del(ctx context.Context, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(s.bucket)
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(key))
	})
}

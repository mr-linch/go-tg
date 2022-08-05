package session

import "context"

type Store interface {
	// Set saves a session session data.
	Set(ctx context.Context, key string, value []byte) error

	// Get returns a session data.
	// If the session data is not found, returns nil and nil.
	Get(ctx context.Context, key string) ([]byte, error)

	// Del deletes a session data.
	Del(ctx context.Context, key string) error
}

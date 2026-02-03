package session

import "context"

// Store define interface for session persistence.
// All stores should have read, write and delete methods.
// See [StoreMemory] for example.
type Store interface {
	// Set saves a session session data.
	Set(ctx context.Context, key string, value []byte) error

	// Get returns a session data.
	// If the session data is not found, returns nil and nil.
	Get(ctx context.Context, key string) ([]byte, error)

	// Del deletes a session data.
	Del(ctx context.Context, key string) error
}

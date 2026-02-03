package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// StoreFile is a session store that stores sessions in files.
type StoreFile struct {
	dir       string
	perms     os.FileMode
	transform func(string) []string
}

var _ Store = (*StoreFile)(nil)

// StoreFileOption is a function that can be passed to NewStoreFile
// to customize the behavior of the store.
type StoreFileOption func(*StoreFile)

// WithStoreFilePerm sets the permissions of the files created by the store.
func WithStoreFilePerm(perms os.FileMode) StoreFileOption {
	return func(store *StoreFile) {
		store.perms = perms
	}
}

// WithStoreFileTransform sets the transform function that is used to
func WithStoreFileTransform(transform func(string) []string) StoreFileOption {
	return func(store *StoreFile) {
		store.transform = transform
	}
}

// NewStoreFile creates a new StoreFile.
func NewStoreFile(dir string, opts ...StoreFileOption) *StoreFile {
	store := &StoreFile{
		dir:   dir,
		perms: 0o666,
		transform: func(key string) []string {
			return []string{key}
		},
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

func (store *StoreFile) getSessionPath(key string) string {
	paths := append([]string{store.dir}, store.transform(key)...)

	return filepath.Join(paths...) + ".session"
}

// Set stores the session data.
func (store *StoreFile) Set(ctx context.Context, key string, value []byte) error {
	path := store.getSessionPath(key)

	if err := store.ensureDirExists(path); err != nil {
		return fmt.Errorf("create dir if not exists: %w", err)
	}

	if err := os.WriteFile(path, value, store.perms); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func (store *StoreFile) ensureDirExists(filePath string) error {
	parent := filepath.Dir(filePath)

	return os.MkdirAll(parent, os.ModePerm)
}

// Get retrieves the session data.
func (store *StoreFile) Get(ctx context.Context, key string) ([]byte, error) {
	path := store.getSessionPath(key)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return data, nil
}

// Del deletes the session data.
func (store *StoreFile) Del(ctx context.Context, key string) error {
	path := store.getSessionPath(key)

	if err := os.Remove(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("remove file: %w", err)
	}

	return nil
}

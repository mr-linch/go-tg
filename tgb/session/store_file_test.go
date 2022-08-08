package session

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreFile_New(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-store-file-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	t.Run("Default", func(t *testing.T) {
		store := NewStoreFile(dir)

		assert.Equal(t, dir, store.dir)
		assert.Equal(t, os.FileMode(0666), store.perms)
		assert.Equal(t, []string{"abc"}, store.transform("abc"))
	})

	t.Run("Custom", func(t *testing.T) {
		store := NewStoreFile(dir,
			WithStoreFilePerm(os.FileMode(0644)),
			WithStoreFileTransform(func(key string) []string {
				return strings.Split(key, "_")
			}),
		)

		assert.Equal(t, dir, store.dir)
		assert.Equal(t, os.FileMode(0644), store.perms)
		assert.Equal(t, []string{"a", "b"}, store.transform("a_b"))
	})
}

func TestStoreFile_Set(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-store-file-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	store := NewStoreFile(dir, WithStoreFileTransform(func(key string) []string {
		return strings.Split(key, "_")
	}))

	err = store.Set(context.Background(), "k_e_y", []byte("value"))
	assert.NoError(t, err)

	f, err := os.Open(filepath.Join(dir, "k", "e", "y.session"))
	assert.NoError(t, err)
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, []byte("value"), b)
}

func TestFileStore_Generic(t *testing.T) {
	dir, err := os.MkdirTemp("", "session-store-file-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	genericStoreTest(t, NewStoreFile(dir))
}

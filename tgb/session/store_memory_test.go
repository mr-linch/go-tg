package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	store := NewStoreMemory()

	err := store.Set(context.Background(), "key", []byte("value"))
	require.NoError(t, err)

	v, err := store.Get(context.Background(), "key")
	require.NoError(t, err)
	require.Equal(t, []byte("value"), v)

	err = store.Del(context.Background(), "key")
	require.NoError(t, err)

	v, err = store.Get(context.Background(), "key")
	require.NoError(t, err)
	require.Nil(t, v)
}

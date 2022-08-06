package e2e

import (
	"context"
	"testing"

	"github.com/mr-linch/go-tg/tgb/session"
	"github.com/stretchr/testify/require"
)

func logic(t *testing.T, store session.Store) {
	t.Helper()

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

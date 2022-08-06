//go:build e2e
// +build e2e

package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/mr-linch/go-tg/tgb/session/sessionsql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func logic(t *testing.T, db *sql.DB, queries sessionsql.Queries) {
	t.Helper()

	store := sessionsql.New(db, "session", queries)

	assert.NoError(t, store.Setup(context.Background()))

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

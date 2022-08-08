//go:build e2e
// +build e2e

package tests

import (
	"database/sql"
	"testing"

	"github.com/mr-linch/go-tg/tgb/session/sessionsql"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestSQLite3(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	assert.NoError(t, err, "db init")

	logic(t, db, sessionsql.SQLite3)
}

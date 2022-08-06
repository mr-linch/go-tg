//go:build e2e
// +build e2e

package tests

import (
	"database/sql"
	"os"
	"testing"

	"github.com/mr-linch/go-tg/tgb/session/sessionsql"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

func TestPostgreSQL(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")

	if dsn == "" {
		t.Skip("skip because POSTGRES_DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	assert.NoError(t, err, "open db")
	defer db.Close()

	db.Exec("drop table if exists session")

	logic(t, db, sessionsql.PostgreSQL)
}

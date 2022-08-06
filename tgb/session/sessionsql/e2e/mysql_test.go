//go:build e2e
// +build e2e

package tests

import (
	"database/sql"
	"os"
	"testing"

	"github.com/mr-linch/go-tg/tgb/session/sessionsql"
	"github.com/stretchr/testify/assert"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQL(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")

	if dsn == "" {
		t.Skip("skip because MYSQL_DSN is not set")
	}

	t.Logf("MYSQL_DSN: %s", dsn)

	db, err := sql.Open("mysql", dsn)
	assert.NoError(t, err, "open db")
	defer db.Close()

	db.Exec("drop table if exists `session`")

	logic(t, db, sessionsql.MySQL)
}

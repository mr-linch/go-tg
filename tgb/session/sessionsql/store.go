package sessionsql

import (
	"context"
	"database/sql"
	"fmt"
)

// Queries is a set of queries for session table.
type Queries struct {
	// Setup is a query for creating table for session.
	Setup string
	// Set is a query for setting a session value.
	Set string
	// Get is a query for getting a session value.
	Get string
	// Del is a query for deleting a session value.
	Del string
}

var (
	// SQLite3 is a set of queries for SQLite3.
	SQLite3 = Queries{
		Setup: `create table if not exists "%s" (
			key text primary key not null,
			value blob not null,
			created_at timestamp not null default current_timestamp,
			updated_at timestamp
		)`,
		Set: `
			insert into "%s" (key, value) values (?, ?)
			on conflict (key) 
				do update 
					set value = excluded.value, 
					updated_at = current_timestamp;
		`,
		Get: `
			select value 
			from "%s" 
			where key = ?
		`,
		Del: `
			delete from "%s"
			where key = ?
		`,
	}

	// PostgreSQL is a set of queries for PostgreSQL.
	PostgreSQL = Queries{
		Setup: `create table if not exists "%s" (
			key text primary key not null,
			value bytea not null,
			created_at timestamp not null default now(),
			updated_at timestamp
		)`,
		Set: `
			insert into "%s" (key, value) values ($1, $2)
			on conflict (key) 
				do update 
					set value = excluded.value, 
					updated_at = current_timestamp;
		`,
		Get: `
			select value 
			from "%s" 
			where key = $1
		`,
		Del: `
			delete from "%s"
			where key = $1
		`,
	}

	// MySQL is a set of queries for M.
	MySQL = Queries{
		Setup: "create table `%s` (" +
			"`key` varchar(255) primary key not null," +
			"`value` text not null," +
			"`created_at` timestamp not null default current_timestamp," +
			"`updated_at` timestamp" +
			")",
		Set: "insert into `%s` (`key`, `value`) " +
			"values (?, ?) " +
			"on duplicate key update `value` = values(`value`), updated_at = current_timestamp;",
		Get: "select `value` from `%s` where `key` = ?",
		Del: "delete from `%s` where `key` = ?",
	}
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Store struct {
	db      DB
	queries Queries
	table   string
}

func New(db DB, table string, queries Queries) *Store {
	return &Store{
		db:      db,
		table:   table,
		queries: queries,
	}
}

// Setup creates table for session.
func (s *Store) Setup(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(s.queries.Setup, s.table),
	)

	return err
}

// Set sets a session value.
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(s.queries.Set, s.table),
		key, value,
	)

	return err
}

// Get gets a session value.
func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte

	if err := s.db.QueryRowContext(ctx,
		fmt.Sprintf(s.queries.Get, s.table),
		key,
	).Scan(&value); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("query row: %w", err)
	}

	return value, nil
}

// Del deletes a session value.
func (s *Store) Del(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(s.queries.Del, s.table),
		key,
	)

	return err
}

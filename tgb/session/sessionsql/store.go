// Package sessionsql implements a session store using SQL.
//
// To interact with the database, the standard [database/sql] is used.
// Does not import or register any drivers, the user has to do it himself.
// Since database queries are not cross-platform, you need to explicitly specify which set to use.
//
// For example, if you are using PostgreSQL, you should manually init DB and pass the corresponding query set:
//
//  // don't forget to import driver
//  db, err := sql.Open("postgres", "...")
//  if err != nil {
//    return err
//  }
//  defer db.Close()
//
//  store := sessionsql.New(db, "session", sessionsql.PostgreSQL)
//
// See Variables section for more list of built-in query sets.
//
// Session can be created manually or automatically by calling Setup method (uses create table if not exisits):
//
//  if err := store.Setup(ctx); err != nil {
//    return err
//  }
//
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
	// Uses ? as placeholder for parameters.
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
	// Uses $1, $2, $3, ... as placeholder for parameters.
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
	// Uses ? as placeholder for parameters.
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

// DB contains subset of database/sql.DB methods.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Store implements session store backed by SQL DB.
type Store struct {
	db      DB
	queries Queries
	table   string
}

// New creates a new session store backed by SQL DB.
// Table argument is a name of table to use.
// Queries argument is a set of queries for session table.
func New(db DB, table string, queries Queries) *Store {
	return &Store{
		db:      db,
		table:   table,
		queries: queries,
	}
}

// Setup creates table for session in DB.
func (s *Store) Setup(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(s.queries.Setup, s.table),
	)

	return err
}

// Set saves a session in DB
func (s *Store) Set(ctx context.Context, key string, value []byte) error {
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(s.queries.Set, s.table),
		key, value,
	)

	return err
}

// Get gets a session from DB by key.
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

// Del deletes a session in DB.
func (s *Store) Del(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf(s.queries.Del, s.table),
		key,
	)

	return err
}

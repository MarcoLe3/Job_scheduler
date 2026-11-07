package repo

import (
	"context"
	"testing"

	"scheduler/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestRepo(t *testing.T) *PostgresRepo {
	t.Helper()

	cfg := db.DSNConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "test",
		Password: "test",
		Dbname:   "test",
	}

	db, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		t.Skipf("test database not reachable: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return NewPostgresRepo(db)
}

func TestInsertRepo(t *testing.T) {

}

func TestGetRepo(t *testing.T) {
	
}
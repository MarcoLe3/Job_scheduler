package repo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"scheduler/internal/db"
	"scheduler/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestRepo(t *testing.T) *PostgresRepo {
	t.Helper()

	cfg := db.DSNConfig{
		Host:     "localhost",
		Port:     "5342",
		User:     "test",
		Password: "test",
		Dbname:   "test",
	}

	db, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		t.Skipf("test database config invalid: %v", err)
	}

	if err = db.Ping(context.Background()); err != nil {
		t.Skipf("test database not reachable: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return NewPostgresRepo(db)
}

func TestInsertRepo(t *testing.T) {
	repo := newTestRepo(t)
	jobReq := model.JobRequest{
		Type:           "echo",
		Payload:        json.RawMessage(`{"msg": "hello"}`),
		Priority:       1,
		IdempotencyKey: "abc",
		RunAt:          time.Now().UTC().String(),
	}
	job := model.NewJob(jobReq)
	if err := repo.Insert(context.Background(), job); err != nil {
		t.Fatalf("Failed to insert job: %v", err)
	}

	j, err := repo.Get(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if j.ID != job.ID {
		t.Fatalf("Failed to get the same job: %v", err)
	}
}

func TestGetRepo(t *testing.T) {
	repo := newTestRepo(t)
	jobReq := model.JobRequest{
		Type:           "echo",
		Payload:        json.RawMessage(`{"msg": "hello"}`),
		Priority:       1,
		IdempotencyKey: "abd",
		RunAt:          time.Now().UTC().String(),
	}

	job := model.NewJob(jobReq)
	if err := repo.Insert(context.Background(), job); err != nil {
		t.Fatalf("Failed to insert job: %v", err)
	}

	j, err := repo.Get(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if j.ID != job.ID {
		t.Fatal("Failed to get the right job")
	}
}

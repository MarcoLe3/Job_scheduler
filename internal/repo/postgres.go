package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"scheduler/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	DB *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		DB: db,
	}
}

func (r *PostgresRepo) Insert(ctx context.Context, job *model.Job) error {
	_, err := r.DB.Exec(ctx, `
        INSERT INTO jobs (id, idempotency_key, type, payload, status,
                          priority, max_attempts, run_at)
        VALUES ($1, NULLIF($2,''), $3, $4, 'queued', $5, $6, $7)`,
		job.ID, job.IdempotencyKey, job.Type, job.Payload,
		job.Priority, job.MaxAttempts, job.RunAt)
	return err
}

func scanJob(row pgx.Row) (*model.Job, error) {
	var j model.Job
	err := row.Scan(&j.ID, &j.IdempotencyKey, &j.Type, &j.Payload, &j.Status,
		&j.Priority, &j.Attempts, &j.MaxAttempts,
		&j.RunAt, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &j, err
}

func (r *PostgresRepo) Get(ctx context.Context, id string) (*model.Job, error) {
	job, err := scanJob(r.DB.QueryRow(ctx, `SELECT * FROM jobs WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	return job, nil
}

func (r *PostgresRepo) MarkDone(ctx context.Context, id string, updTime time.Time) error {
	tag, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'done' WHERE id = $1`, id)
	if err != nil {
		return nil
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("job %s not found and was not marked done", id)
	}
	return err
}

func (r *PostgresRepo) MarkDead(ctx context.Context, id string, errMsg string, updTime time.Time) error {
	tag, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'dead' WHERE id = $1`, id)
	if err != nil {
		return nil
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("job %s not found and was not marked dead", id)
	}
	return err
}

func (r *PostgresRepo) MarkFailed(ctx context.Context, id string, errMsg string, updTime time.Time) error {
	tag, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'done', updated_at = $1, last_error = $2 WHERE id = $3`, updTime, errMsg, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("job %s not found and was not marked running", id)
	}

	return err
}

func (r *PostgresRepo) MarkRunning(ctx context.Context, id string, updTime time.Time) error {
	tag, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'running', updated_at = $1 WHERE id = $2`, updTime, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("job %s not found and was not marked running", id)
	}
	return err
}

func (r *PostgresRepo) Requeue(ctx context.Context, id string, errMsg string, t time.Duration, updTime time.Time) error {
	tag, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'queued', attempts = attempts + 1, updated_at = $1, run_at = $2, last_error = $3 WHERE id = $4 and status = 'running'`, updTime, time.Now().UTC().Add(t), errMsg, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("requeue %s: job not found or not running", id)
	}
	return err
}

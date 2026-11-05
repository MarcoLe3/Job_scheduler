package repo

import (
	"context"
	"errors"

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

func (r *PostgresRepo) MarkDone(ctx context.Context, id string) error {
	_, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'done' WHERE id = $1`, id)
	return err
}

func (r *PostgresRepo) MarkDead(ctx context.Context, id string) error {
	_, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'dead' WHERE id = $1`, id)
	return err
}

func (r *PostgresRepo) MarkFailed(ctx context.Context, id string) error {
	_, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'done' WHERE id = $1`, id)
	return err
}

func (r *PostgresRepo) MarkRunning(ctx context.Context, id string) error {
	_, err := r.DB.Exec(ctx, `UPDATE jobs SET status = 'running' WHERE id = $1`, id)
	return err
}

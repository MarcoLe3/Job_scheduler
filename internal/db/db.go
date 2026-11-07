package db

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DSNConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

func (d DSNConfig) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     net.JoinHostPort(d.Host, d.Port),
		Path:     d.Dbname,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func DSNConfigFromEnv() (DSNConfig, error) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return DSNConfig{}, fmt.Errorf("database connection details are not fully set")
	}

	c := DSNConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Dbname:   dbname,
	}

	return c, nil
}

func OpenDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"scheduler/internal/cache"
	"scheduler/internal/db"
	"scheduler/internal/model"
	"scheduler/internal/repo"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load(); err != nil {
		log.Println("no .env loaded or found")
	}

	cfg, err := db.DSNConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := db.OpenDB(ctx, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	pgRepo := repo.NewPostgresRepo(db)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	if err = rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis fail %v", err)
	}
	defer rdb.Close()

	cache := cache.NewCache(rdb, time.Hour*2)
	queue := NewBlockingQueue[*model.Job](1000)
	svc := NewService(pgRepo, cache, queue)

	pool := NewWorkerPool(queue, pgRepo)
	pool.AddRequest("echo", EchoRequest)
	pool.AddRequest("stripe", StripeRequest)

	pool.Start(ctx, 8)
	app := NewHandler(svc)
	srv := &http.Server{Addr: ":4000", Handler: app.routes()}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("%v", err)
		}
	}()

	log.Println("Listening on port :4000")

	<-ctx.Done()
	log.Println("shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	srv.Shutdown(shutdownCtx)
	pool.Stop()
}

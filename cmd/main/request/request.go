package request

import (
	"context"
	"fmt"
	"log"

	"scheduler/internal/model"
)

type Request func(ctx context.Context, j *model.Job) error

func EchoRequest(ctx context.Context, j *model.Job) error {
	log.Printf("echo: %s", j.Payload)
	return nil
}

func StripeRequest(ctx context.Context, j *model.Job) error {
	log.Print("sent request to stripe")
	return nil
}

func FlakyRequest(ctx context.Context, j *model.Job) error {
	if j.Attempts < j.MaxAttempts {
		return fmt.Errorf("simulate failure on attempt %v: attempt: %d", j.ID, j.Attempts)
	}
	return nil
}

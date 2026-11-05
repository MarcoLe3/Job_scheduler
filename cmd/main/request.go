package main

import (
	"context"
	"log"

	"scheduler/internal/model"
)

type Request func (ctx context.Context, j *model.Job) error 

func EchoRequest(ctx context.Context, j *model.Job) error {
	log.Printf("echo: %s",j.Payload)
	return nil
}

func StripeRequest(ctx context.Context, j *model.Job) error {
	log.Print("sent request to stripe")
	return nil
}

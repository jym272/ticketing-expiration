package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
	"workspace/nats"
)

type OrderProcessor struct {
	subject string
}

func OrderCreated(order nats.OrdersCreated) (*asynq.Task, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(nats.OrderCreated, payload), nil
}

func (processor *OrderProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p nats.OrdersCreated
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Order Id: src=%d", p.ID)
	log.Printf("Order ExpiresAt: src=%s", p.ExpiresAt)
	return nil
}

func NewOrderProcessor(subj string) *OrderProcessor {
	return &OrderProcessor{
		subject: subj,
	}
}

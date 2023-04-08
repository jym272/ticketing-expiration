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

	switch processor.subject {
	case nats.OrderCreated: // the thas has been procesed, is time to cancel the order, is expira alredy
		var p nats.OrdersCreated
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}
		log.Printf("Order Id: src=%d", p.ID)
		log.Printf("Order ExpiresAt: src=%s", p.ExpiresAt)
		// publish an order created
		nats.Publish(nats.OrderCreated, p)
	default:
		return fmt.Errorf("json.Unmarshal failed: %v: %w", "unknown subject", asynq.SkipRetry)
	}
	return nil
}

func NewOrderProcessor(subj string) *OrderProcessor {
	return &OrderProcessor{
		subject: subj,
	}
}

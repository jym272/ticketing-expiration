package async

import (
	"context"
	"encoding/json"
	"fmt"
	"workspace/nats"

	log "github.com/sirupsen/logrus"

	"github.com/hibiken/asynq"
)

type Processor struct {
	subject nats.Subject
}

func CreateTask[T any](msg T, subject nats.Subject) (*asynq.Task, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(string(subject), payload), nil
}

func (processor *Processor) ProcessTask(_ context.Context, t *asynq.Task) error {
	switch processor.subject {
	case nats.OrderCreated:
		var p nats.OrdersCreated
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
		}

		l := log.WithFields(log.Fields{
			"subject":    nats.OrderCreated,
			"order_id":   p.ID,
			"expires_at": p.ExpiresAt,
		})
		l.Info("Task processed")

		err := nats.Publish(nats.ExpirationComplete, p)
		if err != nil {
			return fmt.Errorf("publishing to subject=%s has failed", nats.ExpirationComplete)
		}
	case nats.OrderUpdated, nats.ExpirationComplete:
		return fmt.Errorf("the Subject has not yet been implemented: subject=%s: %w", processor.subject, asynq.SkipRetry)
	}

	return nil
}

func NewProcessor(subj nats.Subject) *Processor {
	return &Processor{
		subject: subj,
	}
}

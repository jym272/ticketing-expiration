package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
	"time"
)

// A list of task types.
const (
	TypeEmailDelivery = "email:deliver"
	TypeImageResize   = "image:resize"
)

type EmailDeliveryPayload struct {
	UserID     int
	TemplateID string
}

type ImageResizePayload struct {
	SourceURL string
}
type OrderPayload struct {
	ID        int    `json:"id"`
	ExpiresAt string `json:"expiresAt"`
}

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

//func OrderCreated(orderId int) (*asynq.Task, error) {
//	payload, err := json.Marshal(OrderPayload{OrderID: orderId})
//	if err != nil {
//		return nil, err
//	}
//	return asynq.NewTask(nats.OrderCreated, payload), nil
//}
func NewEmailDeliveryTask(userID int, tmplID string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailDeliveryPayload{UserID: userID, TemplateID: tmplID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEmailDelivery, payload), nil
}

func NewImageResizeTask(src string) (*asynq.Task, error) {
	payload, err := json.Marshal(ImageResizePayload{SourceURL: src})
	if err != nil {
		return nil, err
	}
	// task options can be passed to NewTask, which can be overridden at enqueue time.
	return asynq.NewTask(TypeImageResize, payload, asynq.MaxRetry(5), asynq.Timeout(20*time.Minute)), nil
}

//---------------------------------------------------------------
// Write a function HandleXXXTask to handle the input task.
// Note that it satisfies the asynq.HandlerFunc interface.
//
// Handler doesn't need to be a function. You can define a type
// that satisfies asynq.Handler interface. See examples below.
//---------------------------------------------------------------

func HandleEmailDeliveryTask(ctx context.Context, t *asynq.Task) error {
	var p EmailDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Sending Email to User: user_id=%d, template_id=%s", p.UserID, p.TemplateID)
	// Email delivery code ...
	return nil
}

// ImageProcessor implements asynq.Handler interface.
type ImageProcessor struct {
	// ... fields for struct
}

type OrderProcessor struct {
	// ... fields for struct
	subject string
}

func (processor *ImageProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p ImageResizePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Resizing image: src=%s", p.SourceURL)
	// Image resizing code ...
	return nil
}

func (processor *OrderProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p OrderPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	log.Printf("Order Id: src=%d", p.ID)
	log.Printf("Order ExpiresAt: src=%s", p.ExpiresAt)
	// Image resizing code ...
	return nil
}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

func NewOrderProcessor(subj string) *OrderProcessor {
	return &OrderProcessor{
		subject: subj,
	}
}

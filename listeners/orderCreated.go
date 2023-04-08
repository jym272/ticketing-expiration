package listeners

import (
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/nats-io/nats.go"
	"log"
	"time"
	"workspace/async1"
	nt "workspace/nats"
)

type OrdersCreated struct {
	ID        int    `json:"id"`
	ExpiresAt string `json:"expiresAt"`
}

type Message struct {
	OrdersCreated OrdersCreated `json:"orders.created"`
}

func OrderCreatedListener(m *nats.Msg) {
	var orderCreated Message
	err := json.Unmarshal(m.Data, &orderCreated)
	if err != nil {
		fmt.Println("Error unmarshalling message.", err)
		panic(err)
	}
	fmt.Println("OrderCreated received:", string(m.Data), orderCreated)
	// create task
	payload, err := json.Marshal(OrdersCreated{ID: orderCreated.OrdersCreated.ID, ExpiresAt: orderCreated.OrdersCreated.ExpiresAt})
	if err != nil {
		fmt.Println("Error marshalling task payload.", err)
		panic(err)
	}
	task := asynq.NewTask(nt.OrderCreated, payload)

	taskInfo, err := async1.EnqueueOrder(task, orderCreated.OrdersCreated.ExpiresAt)

	log.Printf("enqueued task: id=%s queue=%s", taskInfo.ID, taskInfo.Queue)

	if err != nil {
		fmt.Println("Error enqueuing task.", err)
		m.NakWithDelay(2 * time.Second)
		return
	}
	m.Ack()
}

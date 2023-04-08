package listeners

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"workspace/async1"
	"workspace/async1/tasks"
	nt "workspace/nats"
)

type Message struct {
	OrdersCreated nt.OrdersCreated `json:"orders.created"`
}

func OrderCreated(m *nats.Msg) {
	var message Message

	err := json.Unmarshal(m.Data, &message)
	if err != nil {
		log.Fatalf("Error unmarshalling message: %v", err)
	}
	newOrder := message.OrdersCreated
	fmt.Println("OrderCreated received:", string(m.Data), message)

	task, err := tasks.OrderCreated(newOrder)
	if err != nil {
		log.Fatalf("Error creating task: %v", err)
	}

	taskInfo, err := async1.EnqueueOrder(task, newOrder.ExpiresAt)
	if err != nil {
		log.Fatalf("Error enqueuing task: %v", err)
	}

	log.Printf("enqueued task: id=%s queue=%s", taskInfo.ID, taskInfo.Queue)
	m.Ack()
	//if err != nil {
	//	fmt.Println("Error enqueuing task.", err)
	//	m.NakWithDelay(2 * time.Second)
	//	return
	//}
}

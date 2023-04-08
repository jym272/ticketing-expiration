package listeners

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"log"
	"time"
	"workspace/async"
	"workspace/async/tasks"
	nt "workspace/nats"
)

type Message struct {
	OrdersCreated nt.OrdersCreated `json:"orders.created"`
}

func OrderCreated(m *nats.Msg) {
	var message Message

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	logger.Info("ORDER CREATED")

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

	expiresAt, err := time.Parse(time.RFC3339, newOrder.ExpiresAt)
	if err != nil {
		log.Fatalf("could not parse time: %v", err)
	}

	taskInfo, err := async.EnqueueOrder(task, expiresAt)
	if err != nil {
		log.Fatalf("Error enqueuing task: %v", err)
	}

	logrus.Debugln("enqueued task: id=%s queue=%s", taskInfo.ID, taskInfo.Queue)
	//fmt.Printf("enqueued task: id=%s queue=%s", taskInfo.ID, taskInfo.Queue)
	m.Ack()
	//if err != nil {
	//	fmt.Println("Error enqueuing task.", err)
	//	m.NakWithDelay(2 * time.Second)
	//	return
	//}
}

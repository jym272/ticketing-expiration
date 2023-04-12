package callbacks

import (
	"encoding/json"
	"time"
	"workspace/async"
	nt "workspace/nats"

	"github.com/hibiken/asynq"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	OrdersCreated nt.OrdersCreated `json:"orders.created"`
}

const NackDelay = 1000 * time.Millisecond
const MaxRetries = 3

func nakTheMsg(m *nats.Msg) error {
	metadata, err := m.Metadata()
	if err != nil {
		log.Error("Error getting metadata")
		return err
	}

	log.Infof("Number of deliveries: %d", metadata.NumDelivered)

	if int(metadata.NumDelivered) >= MaxRetries {
		log.Infof("Max retries reached %d, terminating message", MaxRetries)

		err = m.Term()

		if err != nil {
			log.Error("Error term the message")
			return err
		}

		return nil
	}

	err = m.NakWithDelay(NackDelay)
	if err != nil {
		log.Error("Error nacking the message")
		return err
	}

	return nil
}

func OrderCreated(m *nats.Msg) {
	var message Message

	if err := json.Unmarshal(m.Data, &message); err != nil {
		log.Errorf("Error unmarshalling message: %v", err)
		err = m.Term()

		if err != nil {
			log.Error("Error term the message")
			return
		}

		return
	}

	order := message.OrdersCreated

	fields := log.Fields{
		"subject": nt.OrderCreated,
		"data":    string(m.Data),
	}
	l := log.WithFields(fields)
	l.Info("Message received")

	expiresAt, err := time.Parse(time.RFC3339, order.ExpiresAt)
	if err != nil {
		l.Errorf("could not parse time: %v", err)
		err = m.Term()

		if err != nil {
			l.Error("Error term the message")
			return
		}

		return
	}

	task, err := async.CreateTask(order, nt.OrderCreated)
	if err != nil {
		l.Errorf("Error creating task: %v", err)
		err = nakTheMsg(m)

		if err != nil {
			l.Error("Error nacking the message")
			return
		}

		return
	}

	taskInfo, err := async.EnqueueTask(task, asynq.ProcessAt(expiresAt))
	if err != nil {
		l.Errorf("Error enqueuing task: %v", err)
		err = nakTheMsg(m)

		if err != nil {
			l.Error("Error nacking the message")
			return
		}

		return
	}

	fields["task_id"] = taskInfo.ID
	fields["task_queue"] = taskInfo.Queue
	l = log.WithFields(fields)
	l.Info("Enqueued task")

	err = m.Ack()
	if err != nil {
		l.Errorf("Error ack: %v", err)
		err = m.Term()

		if err != nil {
			l.Error("Error term the message")
			return
		}

		return
	}
}

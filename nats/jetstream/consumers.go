package jetstream

import (
	"github.com/nats-io/nats.go"

	"fmt"
)

func FindConsumer(js nats.JetStreamContext, stream, durableName string) bool {
	for consumerInfo := range js.Consumers(stream) {
		if consumerInfo.Config.Durable == durableName {
			return true
		}
	}
	return false
}

func CreateConsumer(js nats.JetStreamContext, stream, durableName, queueGroupName, filterSubject string) {
	_, err := js.AddConsumer(stream, &nats.ConsumerConfig{
		Durable:        durableName,
		DeliverPolicy:  nats.DeliverAllPolicy,
		AckPolicy:      nats.AckExplicitPolicy,
		DeliverSubject: nats.NewInbox(),
		DeliverGroup:   queueGroupName,
		FilterSubject:  filterSubject,
	})
	if err != nil {
		fmt.Println("Error creating consumer.", err)
		panic(err)
	}
}

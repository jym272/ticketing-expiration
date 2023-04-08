package nats

import (
	"fmt"
	"time"
	"workspace/nats/utils"

	"github.com/nats-io/nats.go"
)

const TIMEOUT = 5 * time.Second

func Subscribe(subject string, cb nats.MsgHandler) {
	js := GetInstance().Js
	sub, err := js.QueueSubscribeSync(subject, subject,
		nats.Bind(utils.ExtractStreamName(subject), utils.GetDurableName(subject)),
		nats.ManualAck(),
	)

	if err != nil {
		fmt.Println("Error subscribing to subject.", err)
		panic(err)
	}

	fmt.Println("Subscribed to subject:", subject)

	for {
		msg, err := sub.NextMsg(TIMEOUT)
		if err != nil {
			if err == nats.ErrTimeout {
				fmt.Println("Timeout waiting for message.")
				continue
			}
			// Because of MaxReconnectAttempts, probably nats: connection closed
			// is panic-worthy. -> this goroutine will die and the process will exit.
			fmt.Println("Error getting next message.", err)
			panic(err)
		}

		fmt.Println("Message received:", msg.Subject, string(msg.Data))
		cb(msg)
	}
}

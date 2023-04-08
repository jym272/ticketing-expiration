package nats

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"time"
	"workspace/nats/utils"
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
		// TODO: aca tengo que refactorizar, quizas no tenega que monitorrear
		msg, err := sub.NextMsg(TIMEOUT)
		if err != nil {
			if err == nats.ErrTimeout {
				fmt.Println("Timeout waiting for message.")
				continue
			}

			fmt.Println("Error getting next message.", err)
			panic(err)
		}

		fmt.Println("Message received:", msg.Subject, string(msg.Data))
		cb(msg)
	}
}

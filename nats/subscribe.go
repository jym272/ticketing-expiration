package nats

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"time"
	"workspace/nats/utils"
)

func Subscribe(js nats.JetStreamContext, subject string, cb nats.MsgHandler) {
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
		msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			if err == nats.ErrTimeout {
				//fmt.Println("Timeout waiting for message.")
				continue
			}
			fmt.Println("Error getting next message.", err)
			panic(err)
		}
		//   log(`[${m.seq}]: [${sub.getProcessed()}]: ${sc.decode(m.data)}`);
		fmt.Println("Message received:", msg.Subject)
		cb(msg)
	}
}

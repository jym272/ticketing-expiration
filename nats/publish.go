package nats

import (
	"encoding/json"
	"log"
)

func Publish(subject string, msg interface{}) {
	js := GetInstance().Js

	if subject == OrderCreated {
		if _, ok := msg.(OrdersCreated); !ok {
			log.Println("Error casting message:", msg)
			return
		}

		data, err := json.Marshal(msg.(OrdersCreated))

		if err != nil {
			log.Println("Error marshalling message:", err)
			return
		}

		paf, err := js.PublishAsync("orders.cancelled", data)
		if err != nil {
			log.Println("Error publishing message:", err)
			return
		}
		select {
		case errChan := <-paf.Err():
			log.Println("Error publishing message:", errChan)
		case pa := <-paf.Ok():
			log.Printf("Published message to %s, seq: %d, duplicate: %t, domain: %s", pa.Stream, pa.Sequence, pa.Duplicate, pa.Domain)
		}
	}
}

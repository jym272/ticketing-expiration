package nats

import (
	"encoding/json"
	"log"
)

func Publish(subject string, msg interface{}) {
	js := GetInstance().Js

	switch subject {
	case OrderCreated: // TODO: actually is need tyo publush to oher chan
		p := msg.(OrdersCreated)
		data, err := json.Marshal(p)
		if err != nil {
			log.Println("Error marshalling message:", err)
			return
		}
		pa, err := js.PublishAsync("orders.cancelled", data)
		okChan := pa.Ok()
		errChan := pa.Err()
		select {
		case err := <-errChan:
			log.Println("Error publishing message:", err)
		case pa := <-okChan:
			log.Printf("Published message to %s, seq: %d, duplicate: %t, domain: %s", pa.Stream, pa.Sequence, pa.Duplicate, pa.Domain)
		}

	}
}

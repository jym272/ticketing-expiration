package nats

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

func Publish(subject Subject, msg interface{}) {
	l := log.WithFields(log.Fields{
		"subject": subject,
	})
	js := GetInstance().Js

	if subject == OrderCancelled {
		if _, ok := msg.(OrdersCreated); !ok {
			l.Error("Error casting message:", msg)
			return
		}

		data, err := json.Marshal(msg.(OrdersCreated))

		if err != nil {
			l.Error("Error marshalling message:", err)
			return
		}

		paf, err := js.PublishAsync(string(OrderCancelled), data)
		if err != nil {
			l.Error("Error publishing message:", err)
			return
		}
		select {
		case errChan := <-paf.Err():
			l.Error("Error publishing message:", errChan)
		case pa := <-paf.Ok():
			l = log.WithFields(log.Fields{
				"subject": subject,
				"seq":     pa.Sequence,
				"dup":     pa.Duplicate,
				"stream":  pa.Stream,
			})
			l.Info("Message published.")
		}
	}
}

package nats

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

func Publish[T any](subject Subject, msg T) {
	l := log.WithFields(log.Fields{
		"subject": subject,
	})
	js := GetInstance().Js

	toEncode := make(map[Subject]T)
	toEncode[subject] = msg

	data, err := json.Marshal(toEncode)
	if err != nil {
		l.Error("Error marshalling message:", err)
		return
	}

	paf, err := js.PublishAsync(string(subject), data)
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

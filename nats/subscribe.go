package nats

import (
	"time"
	"workspace/nats/utils"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

const TIMEOUT = 5 * time.Second

func Subscribe(subject string, cb nats.MsgHandler) {
	l := log.WithFields(log.Fields{
		"subject": subject,
	})
	js := GetInstance().Js
	sub, err := js.QueueSubscribeSync(subject, subject,
		nats.Bind(utils.ExtractStreamName(subject), utils.GetDurableName(subject)),
		nats.ManualAck(),
	)

	if err != nil {
		l.Panic("Error subscribing to subject", err)
	}

	l.Info("Subscribed")

	for {
		l = log.WithFields(log.Fields{
			"subject": subject,
		})

		msg, err := sub.NextMsg(TIMEOUT)
		if err != nil {
			if err == nats.ErrTimeout {
				l.Trace("Timeout waiting for message")
				continue
			}
			// Because of MaxReconnectAttempts, probably nats: connection closed
			// is panic-worthy. -> this goroutine will die and the process will exit.
			l.Panic("Error getting next message", err)
		}

		l.Info("Message received")
		cb(msg)
	}
}

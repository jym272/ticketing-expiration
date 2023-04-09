package nats

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"

	log "github.com/sirupsen/logrus"
)

type Stream struct {
	name     string
	subjects []string
}

func (s Stream) CreateStream(stream string, subjects []string) Stream {
	validateStream(stream, subjects)
	s.name = stream
	s.subjects = subjects

	return s
}

func createStream(js nats.JetStreamContext, stream string) {
	subj := stream + ".>" // ie: tickets.> -> tickets.one, tickets.one.two

	_, err := js.AddStream(&nats.StreamConfig{
		Name:     stream,
		Subjects: []string{subj}, // any number of tokens->ie: events.one, events.one.two
	})
	if err != nil {
		fmt.Println("Error creating stream.", err)
		panic(err)
	}
}

func findStream(js nats.JetStreamContext, stream string) bool {
	for name := range js.StreamNames() {
		if name == stream {
			return true
		}
	}

	return false
}

func validateStream(stream string, subjects []string) {
	l := log.WithFields(log.Fields{
		"stream":   stream,
		"subjects": subjects,
	})
	if stream == "" {
		l.Panic("stream name cannot be empty")
	}

	if len(subjects) == 0 {
		l.Panic("subjects cannot be empty in stream")
	}

	for _, subject := range subjects {
		if subject == "" {
			l.Panic("subject cannot be empty in stream")
		}

		if !strings.HasPrefix(subject, stream+".") {
			l.Panic("subject does not start with stream.")
		}
	}
}

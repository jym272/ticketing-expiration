package nats

import (
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

type Subject string
type Stream string

const (
	Orders             Stream  = "orders"
	Expiration         Stream  = "expiration"
	OrderCreated       Subject = "orders.created"
	OrderCancelled     Subject = "orders.cancelled"
	ExpirationComplete Subject = "expiration.complete"
)

type OrdersCreated struct {
	ExpiresAt string `json:"expiresAt"`
	ID        int    `json:"id"`
}

type Context struct {
	Nc      *nats.Conn
	Js      nats.JetStreamContext
	streams []Streams
}

type Streams struct {
	name     Stream
	subjects []Subject
}

var instance *Context
var once sync.Once

func GetInstance() *Context {
	once.Do(func() {
		instance = &Context{}
	})

	return instance
}

func (c *Context) AddStreams() {
	expirationStream := Streams{}.CreateStream(Expiration, []Subject{ExpirationComplete})
	ordersStream := Streams{}.CreateStream(Orders, []Subject{OrderCreated, OrderCancelled})
	c.streams = []Streams{expirationStream, ordersStream}
}

const MaxReconnectAttempts = 5

func (c *Context) ConnectToNats() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url, nats.MaxReconnects(MaxReconnectAttempts))
	if err != nil {
		log.Panic("Error connecting to NATS.", err)
	}

	c.Nc = nc
	c.Js, err = nc.JetStream()

	if err != nil {
		log.Panic("Error creating JetStream context.", err)
	}
}

func (c *Context) VerifyStreams() {
	for _, stream := range c.streams {
		name := stream.name
		l := log.WithFields(log.Fields{
			"stream": name,
		})

		if !findStream(c.Js, name) {
			l.Warn("Not found. Creating...")
			createStream(c.Js, name)
			l.Info("Created.")

			continue
		}

		l.Debug("Found.")
	}
}

func (c *Context) VerifyConsumers() {
	for _, stream := range c.streams {
		for _, subject := range stream.subjects {
			props := CreateConsumerProps(subject)
			l := log.WithFields(log.Fields{
				"consumer":       props.durableName,
				"stream":         stream.name,
				"durableName":    props.durableName,
				"queueGroupName": props.queueGroupName,
				"filterSubject":  props.filterSubject,
			})

			if !FindConsumer(c.Js, stream.name, props.durableName) {
				l.Warn("Not found. Creating...")
				CreateConsumer(c.Js, stream.name, props)
				l.Info("Created.")

				continue
			}

			l.Debug("Found.")
		}
	}
}

func (s Streams) CreateStream(stream Stream, subjects []Subject) Streams {
	validateStream(stream, subjects)
	s.name = stream
	s.subjects = subjects

	return s
}

func createStream(js nats.JetStreamContext, stream Stream) {
	subj := string(stream) + ".>" // ie: tickets.> -> tickets.one, tickets.one.two

	_, err := js.AddStream(&nats.StreamConfig{
		Name:     string(stream),
		Subjects: []string{subj}, // any number of tokens->ie: events.one, events.one.two
	})
	if err != nil {
		fmt.Println("Error creating stream.", err)
		panic(err)
	}
}

func findStream(js nats.JetStreamContext, stream Stream) bool {
	for name := range js.StreamNames() {
		if name == string(stream) {
			return true
		}
	}

	return false
}

func validateStream(stream Stream, subjects []Subject) {
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

		if !strings.HasPrefix(string(subject), string(stream)+".") {
			l.Panic("subject does not start with stream.")
		}
	}
}

func ExtractStreamName(subject Subject) string {
	parts := strings.Split(string(subject), ".")
	if len(parts) == 0 {
		panic("Subject is empty")
	}

	stream := parts[0]

	return stream
}

func GetDurableName(subject Subject) string {
	parts := strings.Split(string(subject), ".")
	if len(parts) == 0 {
		panic("Subject is empty")
	}

	upperCaseParts := make([]string, 0, len(parts))

	for _, part := range parts {
		upperCaseParts = append(upperCaseParts, strings.ToUpper(part))
	}

	return strings.Join(upperCaseParts, "_")
}

type ConsumerProps struct {
	durableName    string
	queueGroupName string
	filterSubject  string
}

func CreateConsumerProps(subject Subject) *ConsumerProps {
	return &ConsumerProps{
		durableName:    GetDurableName(subject),
		queueGroupName: string(subject),
		filterSubject:  string(subject),
	}
}

func FindConsumer(js nats.JetStreamContext, stream Stream, durableName string) bool {
	for consumerInfo := range js.Consumers(string(stream)) {
		if consumerInfo.Config.Durable == durableName {
			return true
		}
	}

	return false
}

func CreateConsumer(js nats.JetStreamContext, stream Stream, props *ConsumerProps) {
	_, err := js.AddConsumer(string(stream), &nats.ConsumerConfig{
		Durable:        props.durableName,
		DeliverPolicy:  nats.DeliverAllPolicy,
		AckPolicy:      nats.AckExplicitPolicy,
		DeliverSubject: nats.NewInbox(),
		DeliverGroup:   props.queueGroupName,
		FilterSubject:  props.filterSubject,
	})
	if err != nil {
		fmt.Println("Error creating consumer.", err)
		panic(err)
	}
}

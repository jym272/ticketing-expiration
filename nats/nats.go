package nats

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

type Subject string
type Stream string

const TIMEOUT = 5 * time.Second
const MaxReconnectAttempts = 5

const (
	Orders             Stream  = "orders"
	Expiration         Stream  = "expiration"
	OrderCreated       Subject = "orders.created"
	OrderUpdated       Subject = "orders.updated"
	ExpirationComplete Subject = "expiration.complete"
)

type OrdersCreated struct {
	ExpiresAt string `json:"expiresAt"`
	ID        int    `json:"id"`
}

type Subscriber struct {
	Cb      nats.MsgHandler
	Subject Subject
}
type Nats struct {
	Nc            *nats.Conn
	Js            nats.JetStreamContext
	streams       []Streams
	subscribers   []Subscriber
	done          chan struct{}
	subscriptions []*nats.Subscription
}

type Streams struct {
	Name     Stream
	Subjects []Subject
}

var instance *Nats
var once sync.Once

func GetInstance() *Nats {
	once.Do(func() {
		instance = &Nats{}
	})

	return instance
}

func GetNats(subs []Subscriber, unvalidatedStreams []Streams) *Nats {
	for _, stream := range unvalidatedStreams {
		validateStream(stream.Name, stream.Subjects)
	}

	n := GetInstance()
	n.subscribers = subs
	n.streams = unvalidatedStreams
	n.done = make(chan struct{})

	return n
}

func (n *Nats) StartServer() {
	n.ConnectToNats()
	n.VerifyStreams()
	n.VerifyConsumers()
}
func (n *Nats) subscribe(subject Subject, cb nats.MsgHandler) {
	l := log.WithFields(log.Fields{
		"subject": subject,
	})
	js := n.Js
	sub, err := js.QueueSubscribeSync(string(subject), string(subject),
		nats.Bind(ExtractStreamName(subject), GetDurableName(subject)),
		nats.ManualAck(),
	)

	if err != nil {
		l.Panic("Error subscribing to subject: ", subject, err)
	}

	n.subscriptions = append(n.subscriptions, sub)

	l.Info("Subscribed")

	for {
		msg, err := sub.NextMsg(TIMEOUT)
		if err != nil {
			if err == nats.ErrTimeout {
				l.Trace("Timeout waiting for message")
				continue
			}
			// ErrBadSubscription -> because of unsubscribing/draining, we get this error
			if err == nats.ErrBadSubscription {
				break // stop the sub
			}
			// Because of MaxReconnectAttempts, probably nats: connection closed
			l.Panic("Error getting next message", err)
		}

		cb(msg)
	}
	l.Info("Unsubscribed")
}
func (n *Nats) Start(wg *sync.WaitGroup) {
	n.StartServer()

	for _, subscriber := range n.subscribers {
		wg.Add(1)

		go func(sub Subscriber) {
			defer wg.Done()
			n.subscribe(sub.Subject, sub.Cb)
		}(subscriber)
	}

	go func() {
		for range n.done {
			// for _, sub := range n.subscriptions {
			//	 err := sub.Unsubscribe()
			//	 if err != nil {
			//		 log.Error("Error unsubscribing from nats server", err)
			//		 return
			//	 }
			// }
			// Drain Nc also Drain all subs
			err := n.Nc.Drain()
			if err != nil {
				log.Error("Error draining nats server", err)
				return
			}
		}
	}()
}
func (n *Nats) Shutdown() {
	n.done <- struct{}{}
}

func (n *Nats) ConnectToNats() {
	natsServerHost := os.Getenv("NATS_SERVER_HOST")
	natsServerPort := os.Getenv("NATS_SERVER_PORT")

	if natsServerHost == "" || natsServerPort == "" {
		log.Panic("NATS_SERVER_HOST or NATS_SERVER_PORT is not set")
	}

	url := fmt.Sprintf("nats://%s:%s", natsServerHost, natsServerPort)

	nc, err := nats.Connect(url, nats.MaxReconnects(MaxReconnectAttempts))
	if err != nil {
		log.Panic("Error connecting to NATS.", err)
	}

	n.Nc = nc
	n.Js, err = nc.JetStream()

	if err != nil {
		log.Panic("Error creating JetStream context.", err)
	}
}

func (n *Nats) VerifyStreams() {
	for _, stream := range n.streams {
		name := stream.Name
		l := log.WithFields(log.Fields{
			"stream": name,
		})

		if !findStream(n.Js, name) {
			l.Warn("Not found. Creating...")
			createStream(n.Js, name)
			l.Info("Created.")

			continue
		}

		l.Debug("Found.")
	}
}

func (n *Nats) VerifyConsumers() {
	for _, stream := range n.streams {
		for _, subject := range stream.Subjects {
			props := CreateConsumerProps(subject)
			l := log.WithFields(log.Fields{
				"consumer":       props.durableName,
				"stream":         stream.Name,
				"durableName":    props.durableName,
				"queueGroupName": props.queueGroupName,
				"filterSubject":  props.filterSubject,
			})

			if !FindConsumer(n.Js, stream.Name, props.durableName) {
				l.Warn("Not found. Creating...")
				CreateConsumer(n.Js, stream.Name, props)
				l.Info("Created.")

				continue
			}

			l.Debug("Found.")
		}
	}
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
		"Subjects": subjects,
	})
	if stream == "" {
		l.Panic("stream Name cannot be empty")
	}

	if len(subjects) == 0 {
		l.Panic("Subjects cannot be empty in stream")
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

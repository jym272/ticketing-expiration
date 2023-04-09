package nats

import (
	"os"
	"sync"
	"workspace/nats/jetstream"
	"workspace/nats/utils"

	log "github.com/sirupsen/logrus"

	"github.com/nats-io/nats.go"
)

const (
	Tickets        string = "tickets"
	Orders         string = "orders"
	OrderCreated   string = "orders.created"
	OrderCancelled string = "orders.cancelled"
	TicketCreated  string = "tickets.created"
	TicketUpdated  string = "tickets.updated"
)

type OrdersCreated struct {
	ExpiresAt string `json:"expiresAt"`
	ID        int    `json:"id"`
}

type Context struct {
	Nc      *nats.Conn
	Js      nats.JetStreamContext
	streams []Stream
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
	ticketStream := Stream{}.CreateStream(Tickets, []string{TicketCreated, TicketUpdated})
	ordersStream := Stream{}.CreateStream(Orders, []string{OrderCreated, OrderCancelled})
	c.streams = []Stream{ticketStream, ordersStream}
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
		} else {
			l.Debug("Found.")
		}
	}
}

func (c *Context) VerifyConsumers() {
	for _, stream := range c.streams {
		for _, subject := range stream.subjects {
			durableName, queueGroupName, filterSubject := utils.CreateConsumerProps(subject)
			l := log.WithFields(log.Fields{
				"consumer":       durableName,
				"stream":         stream.name,
				"durableName":    durableName,
				"queueGroupName": queueGroupName,
				"filterSubject":  filterSubject,
			})

			if !jetstream.FindConsumer(c.Js, stream.name, durableName) {
				l.Warn("Not found. Creating...")
				jetstream.CreateConsumer(c.Js, stream.name, durableName, queueGroupName, filterSubject)
				l.Info("Created.")
			} else {
				l.Debug("Found.")
			}
		}
	}
}

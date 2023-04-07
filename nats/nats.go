package main

import (
	"github.com/nats-io/nats.go"
	"strings"

	"fmt"
	"os"
)

const (
	Tickets        string = "tickets"
	Orders         string = "orders"
	OrderCreated   string = "orders.created"
	OrderCancelled string = "orders.cancelled"
	TicketCreated  string = "tickets.created"
	TicketUpdated  string = "tickets.updated"
)

type Stream struct {
	name     string
	subjects []string
}

type Context struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	streams []Stream
}

func validateStream(stream string, subjects []string) {
	if stream == "" {
		panic("stream name cannot be empty")
	}
	if len(subjects) == 0 {
		panic("subjects cannot be empty in stream: " + stream)
	}
	for _, subject := range subjects {
		if subject == "" {
			panic("subject cannot be empty in stream: " + stream)
		}
		if !strings.HasPrefix(subject, stream+".") {
			panic("subject: " + subject + " does not start with stream: " + stream + ".")
		}
	}
}

func (s Stream) CreateStream(stream string, subjects []string) Stream {
	validateStream(stream, subjects)
	s.name = stream
	s.subjects = subjects
	return s
}

func (c *Context) AddStreams() {
	ticketStream := Stream{}.CreateStream(Tickets, []string{TicketCreated, TicketUpdated})
	ordersStream := Stream{}.CreateStream(Orders, []string{OrderCreated, OrderCancelled})
	c.streams = []Stream{ticketStream, ordersStream}
}

func (c *Context) connectToNats() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		fmt.Println("Error connecting to NATS.", err)
		panic(err)
	}

	c.nc = nc
	c.js, err = nc.JetStream()
	if err != nil {
		fmt.Println("Error creating JetStream context.", err)
		panic(err)
	}
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

func (c *Context) verifyStreams() {
	for _, stream := range c.streams {
		name := stream.name
		if !findStream(c.js, name) {
			fmt.Println("Stream not found:", name, "Creating...")
			createStream(c.js, name)
			fmt.Println("Stream created:", name)
		} else {
			fmt.Println("Stream found:", name)
		}
	}
}

func findConsumer(js nats.JetStreamContext, stream string, durableName string) bool {
	for consumerInfo := range js.Consumers(stream) {
		if consumerInfo.Config.Durable == durableName {
			return true
		}
	}
	return false
}

func createConsumer(js nats.JetStreamContext, stream string, durableName string, queueGroupName string, filterSubject string) {
	_, err := js.AddConsumer(stream, &nats.ConsumerConfig{
		Durable:        durableName,
		DeliverPolicy:  nats.DeliverAllPolicy,
		AckPolicy:      nats.AckExplicitPolicy,
		DeliverSubject: nats.NewInbox(),
		DeliverGroup:   queueGroupName,
		FilterSubject:  filterSubject,
	})
	if err != nil {
		fmt.Println("Error creating consumer.", err)
		panic(err)
	}
}

func createConsumerProps(subject string) (string, string, string) {
	parts := strings.Split(subject, ".")
	if len(parts) == 0 {
		panic("Subject is empty")
	}
	var upperCaseParts []string
	for _, part := range parts {
		upperCaseParts = append(upperCaseParts, strings.ToUpper(part))
	}
	durableName := strings.Join(upperCaseParts, "_")
	return durableName, subject, subject
}

func (c *Context) verifyConsumers() {
	for _, stream := range c.streams {
		for _, subject := range stream.subjects {
			durableName, queueGroupName, filterSubject := createConsumerProps(subject)
			if !findConsumer(c.js, stream.name, durableName) {
				fmt.Println("Consumer not found:", durableName, "Creating consumer...")
				createConsumer(c.js, stream.name, durableName, queueGroupName, filterSubject)
				fmt.Println("Consumer created:", durableName)
			} else {
				fmt.Println("Consumer found:", durableName)
			}
		}
	}
}

func main() {
	ctx := Context{}
	ctx.AddStreams()
	ctx.connectToNats()
	ctx.verifyStreams()
	ctx.verifyConsumers()

	defer func(nc *nats.Conn) {
		err := nc.Drain()
		if err != nil {
			fmt.Println("Error draining.", err)
		}
	}(ctx.nc)
}

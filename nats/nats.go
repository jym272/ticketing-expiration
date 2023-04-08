package nats

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"os"
	"workspace/nats/jetstream"
	"workspace/nats/utils"
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
	ID        int    `json:"id"`
	ExpiresAt string `json:"expiresAt"`
}

type Context struct {
	Nc      *nats.Conn
	Js      nats.JetStreamContext
	streams []Stream
}

func (c *Context) AddStreams() {
	ticketStream := Stream{}.CreateStream(Tickets, []string{TicketCreated, TicketUpdated})
	ordersStream := Stream{}.CreateStream(Orders, []string{OrderCreated, OrderCancelled})
	c.streams = []Stream{ticketStream, ordersStream}
}

func (c *Context) ConnectToNats() {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		fmt.Println("Error connecting to NATS.", err)
		panic(err)
	}

	c.Nc = nc
	c.Js, err = nc.JetStream()
	if err != nil {
		fmt.Println("Error creating JetStream context.", err)
		panic(err)
	}
}

func (c *Context) VerifyStreams() {
	for _, stream := range c.streams {
		name := stream.name
		if !findStream(c.Js, name) {
			fmt.Println("Stream not found:", name, "Creating...")
			createStream(c.Js, name)
			fmt.Println("Stream created:", name)
		} else {
			fmt.Println("Stream found:", name)
		}
	}
}

func (c *Context) VerifyConsumers() {
	for _, stream := range c.streams {
		for _, subject := range stream.subjects {
			durableName, queueGroupName, filterSubject := utils.CreateConsumerProps(subject)
			if !jetstream.FindConsumer(c.Js, stream.name, durableName) {
				fmt.Println("Consumer not found:", durableName, "Creating consumer...")
				jetstream.CreateConsumer(c.Js, stream.name, durableName, queueGroupName, filterSubject)
				fmt.Println("Consumer created:", durableName)
			} else {
				fmt.Println("Consumer found:", durableName)
			}
		}
	}
}

package main

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"strings"
	"time"

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

func getDurableName(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) == 0 {
		panic("Subject is empty")
	}
	var upperCaseParts []string
	for _, part := range parts {
		upperCaseParts = append(upperCaseParts, strings.ToUpper(part))
	}
	return strings.Join(upperCaseParts, "_")
}

func createConsumerProps(subject string) (string, string, string) {
	durableName := getDurableName(subject)
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

/*
const getOptsBuilderConfigured = (subj: Subjects): ConsumerOptsBuilder => {
  const opts = consumerOpts();
  opts.queue(subj);
  opts.manualAck();
  opts.bind(extractStreamName(subj), getDurableName(subj));
  return opts;
};

export const subscribe = async (subj: Subjects, cb: (m: JsMsg) => Promise<void>) => {
  if (!js) {
    throw new Error('Jetstream is not defined');
  }
  let sub: JetStreamSubscription;
  try {
    sub = await js.subscribe(subj, getOptsBuilderConfigured(subj));
  } catch (e) {
    log(`Error subscribing to ${subj}`, e);
    throw e;
  }
  log(`[${subj}] subscription opened`);
  for await (const m of sub) {
    log(`[${m.seq}]: [${sub.getProcessed()}]: ${sc.decode(m.data)}`);
    // igual no puedo esperar nada
    void cb(m);
  }
  log(`[${subj}] subscription closed`);
};

*/

/*
  const opts = consumerOpts();
  opts.queue(subj);
  opts.manualAck();
  opts.bind(extractStreamName(subj), getDurableName(subj));
  return opts;
*/
func extractStreamName(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) == 0 {
		panic("Subject is empty")
	}
	stream := parts[0]
	return stream
}

func subscribe(js nats.JetStreamContext, subject string, cb nats.MsgHandler) {
	sub, err := js.QueueSubscribeSync(subject, subject,
		nats.Bind(extractStreamName(subject), getDurableName(subject)),
		nats.ManualAck(),
	)
	if err != nil {
		fmt.Println("Error subscribing to subject.", err)
		panic(err)
	}
	fmt.Println("Subscribed to subject:", subject)
	for {
		msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			if err == nats.ErrTimeout {
				//fmt.Println("Timeout waiting for message.")
				continue
			}
			fmt.Println("Error getting next message.", err)
			panic(err)
		}
		//   log(`[${m.seq}]: [${sub.getProcessed()}]: ${sc.decode(m.data)}`);
		fmt.Println("Message received:", msg.Subject)
		cb(msg)
	}
}

/*
{
  "tickets.created": {
    "id": 44417,
    "title": "New ticket",
    "price": 17
  }
}
*/

type TicketCreatedMessage struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

type Message struct {
	TicketsCreated TicketCreatedMessage `json:"tickets.created"`
}

//type Message struct {
//	TicketsCreated struct {
//		ID    int    `json:"id"`
//		Title string `json:"title"`
//		Price int    `json:"price"`
//	} `json:"tickets.created"`
//}

func main() {
	ctx := Context{}
	ctx.AddStreams()
	ctx.connectToNats()
	ctx.verifyStreams()
	ctx.verifyConsumers()

	// subscribe to TicketCreated
	subscribe(ctx.js, TicketCreated, func(m *nats.Msg) {
		var ticketCreated Message
		err := json.Unmarshal(m.Data, &ticketCreated)
		if err != nil {
			fmt.Println("Error unmarshalling message.", err)
			panic(err)
		}
		fmt.Println("TicketCreated received:", string(m.Data), ticketCreated)
		m.Ack()
	})

	defer func(nc *nats.Conn) {
		err := nc.Drain()
		if err != nil {
			fmt.Println("Error draining.", err)
		}
		fmt.Println("Connection drained.")
	}(ctx.nc)
}

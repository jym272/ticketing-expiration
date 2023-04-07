package main

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	nt "workspace/nats"
	"workspace/nats/listeners"
)

func main() {
	ctx := nt.Context{}
	ctx.AddStreams()
	ctx.ConnectToNats()
	ctx.VerifyStreams()
	ctx.VerifyConsumers()

	// subscribe to TicketCreated
	nt.Subscribe(ctx.Js, nt.TicketCreated, func(m *nats.Msg) {
		var ticketCreated listeners.Message
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
	}(ctx.Nc)
}

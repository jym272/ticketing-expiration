package main

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"syscall"
	nt "workspace/nats"
	"workspace/nats/listeners"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx := nt.Context{}

	ctx.AddStreams()
	ctx.ConnectToNats()
	ctx.VerifyStreams()
	ctx.VerifyConsumers()

	go signalListener(signals, ctx.Nc)

	defer func(nc *nats.Conn) {
		err := nc.Drain()
		if err != nil {
			fmt.Println("Error draining.", err)
		}
		fmt.Println("Connection drained.")
	}(ctx.Nc)

	go nt.Subscribe(ctx.Js, nt.TicketCreated, func(m *nats.Msg) {
		var ticketCreated listeners.Message
		err := json.Unmarshal(m.Data, &ticketCreated)
		if err != nil {
			fmt.Println("Error unmarshalling message.", err)
			panic(err)
		}
		fmt.Println("TicketCreated received:", string(m.Data), ticketCreated)
		m.Ack()
	})

	nt.Subscribe(ctx.Js, nt.TicketUpdated, func(m *nats.Msg) {
		var ticketCreated listeners.Message
		err := json.Unmarshal(m.Data, &ticketCreated)
		if err != nil {
			fmt.Println("Error unmarshalling message.", err)
			panic(err)
		}
		fmt.Println("TicketCreated received:", string(m.Data), ticketCreated)
		m.Ack()
	})

}

func signalListener(signals chan os.Signal, nc *nats.Conn) {
	<-signals
	fmt.Println("Received signal.")
	err := nc.Drain()
	if err != nil {
		fmt.Println("Error draining.", err)
	}
	fmt.Println("Connection drained.")
	os.Exit(0)
}

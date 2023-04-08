package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
	"syscall"
	as "workspace/async1"
	"workspace/listeners"
	nt "workspace/nats"
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
			log.Fatalf("Error draining: %v", err)
		}
		fmt.Println("Connection drained.")
	}(ctx.Nc)

	go as.StartAsyncServer()

	nt.Subscribe(ctx.Js, nt.OrderCreated, listeners.OrderCreatedListener)
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

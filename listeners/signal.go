package listeners

import (
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
)

func Signal(signals chan os.Signal, nc *nats.Conn) {
	<-signals

	err := nc.Drain()

	if err != nil {
		fmt.Println("Error draining.", err)
	}

	fmt.Println("Connection drained.")

	os.Exit(0)
}

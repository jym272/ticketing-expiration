package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"workspace/async"
	cb "workspace/callbacks"

	nt "workspace/nats"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := nt.GetInstance()

	ctx.AddStreams()
	ctx.ConnectToNats()
	ctx.VerifyStreams()
	ctx.VerifyConsumers()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go Signal(signals, ctx.Nc)

	defer func(nc *nats.Conn) {
		err := nc.Drain()
		if err != nil {
			log.Fatal("Error draining", err)
		}

		log.Debug("Connection drained.")
		nc.Close()
	}(ctx.Nc)

	go async.StartServer()
	go nt.Subscribe(nt.OrderCreated, cb.OrderCreated)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker! <3")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "OK"})
	})

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	e.Logger.Fatal(e.Start(":" + httpPort))
}

func Signal(signals chan os.Signal, nc *nats.Conn) {
	<-signals

	err := nc.Drain()

	if err != nil {
		log.Error("Error draining.", err)
	}

	log.Debug("Connection drained.")

	os.Exit(0)
}

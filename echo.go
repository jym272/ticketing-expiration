package main

import (
	stdContext "context"
	"net/http"
	"os"
	"sync"
	"workspace/nats"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

type EchoServer struct {
	server *echo.Echo
	done   chan struct{}
}

// comment
func (e *EchoServer) start(wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		httpPort := os.Getenv("PORT")
		if httpPort == "" {
			log.Panic("PORT env var is not set")
		}

		if err := e.server.Start(":" + httpPort); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		for range e.done {
			if err := e.server.Shutdown(stdContext.Background()); err != nil {
				log.Fatal(err)
			}

			log.Info("EchoServer stopped")
		}
	}()
}
func (e *EchoServer) shutdown() {
	e.done <- struct{}{}
}

func getEcho() *EchoServer {
	e := echo.New()
	if os.Getenv("ECHO_LOGGER") == "true" {
		e.Use(middleware.Logger())
	}

	e.Use(middleware.Recover())

	e.GET("/api/healthz", func(c echo.Context) error {
		n := nats.GetInstance()
		if n.Nc == nil || n.Nc.IsClosed() {
			return c.JSON(http.StatusBadRequest, struct {
				Status string `json:"status"`
			}{Status: "error"})
		}

		return c.JSON(http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "OK"})
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "OK"})
	})

	return &EchoServer{
		server: e,
		done:   make(chan struct{}),
	}
}

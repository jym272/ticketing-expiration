package main

import (
	stdContext "context"
	"net/http"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type EchoServer struct {
	server *echo.Echo
	logger *log.Logger
	done   chan struct{}
}

func (e *EchoServer) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpPort := os.Getenv("PORT")
		if httpPort == "" {
			httpPort = "8080"
		}

		go func() {
			if err := e.server.Start(":" + httpPort); err != nil && err != http.ErrServerClosed {
				e.logger.Fatal(err)
			}
		}()
		for {
			select {
			case <-e.done:
				e.logger.Info("EchoServer stopping")
				if err := e.server.Shutdown(stdContext.Background()); err != nil {
					e.logger.Fatal(err)
				}
				e.logger.Info("EchoServer stopped")
				return
			}
		}
	}()

}

func (e *EchoServer) shutdown() {
	e.done <- struct{}{}
}

func newEcho(logger *log.Logger) *EchoServer {

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
	return &EchoServer{
		server: e,
		done:   make(chan struct{}),
		logger: logger,
	}

}

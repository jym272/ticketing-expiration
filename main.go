package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	as "workspace/async"
	cb "workspace/callbacks"
	nt "workspace/nats"

	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/sys/unix"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type EchoServer struct {
	server *echo.Echo
}

type Server struct {
	wg    sync.WaitGroup
	nats  *nt.Nats
	async *as.Async
	echo  *EchoServer
}

func (srv *Server) waitForSignals() {
	//srv.logger.Info("Send signal TSTP to stop processing new tasks")
	//srv.logger.Info("Send signal TERM or INT to terminate the process")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGTSTP)
	for {
		sig := <-sigs
		if sig == unix.SIGTSTP {
			//srv.async.Stop()
			continue
		}
		break
	}
}

func NewServer() *Server {

	var subs []nt.Subscriber
	sub := &nt.Subscriber{
		Subject: nt.OrderCreated,
		Cb:      cb.OrderCreated,
	}
	subs = append(subs, *sub)
	return &Server{
		nats:  nt.GetNats(subs),
		async: as.GetAsync(),
		echo:  getEcho(),
	}
}

func (srv *Server) Run() error {
	if err := srv.Start(); err != nil {
		return err
	}
	srv.waitForSignals()
	//srv.Shutdown()
	return nil
}

func (e *EchoServer) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpPort := os.Getenv("PORT")
		if httpPort == "" {
			httpPort = "8080"
		}
		e.server.Logger.Fatal(e.server.Start(":" + httpPort))
	}()

}

func (srv *Server) Start() error {

	srv.async.Start(&srv.wg)
	srv.nats.Start(&srv.wg)
	srv.echo.Start(&srv.wg)
	// EL SHUTDOWN usa signals

	return nil
}

func main() {

	srv := NewServer()
	if err := srv.Run(); err != nil {
		log.Fatalf("could not run server: %v", err)
	}

}

func getEcho() *EchoServer {

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
	}

}

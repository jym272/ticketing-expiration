package main

import (
	"os"
	"os/signal"
	"sync"
	as "workspace/async"
	cb "workspace/callbacks"
	nt "workspace/nats"

	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	nats  *nt.Nats
	echo  *EchoServer
	async *as.Async
	wg    sync.WaitGroup
}

func newServer() *Server {
	subs := []nt.Subscriber{
		{Subject: nt.OrderCreated, Cb: cb.OrderCreated},
	}
	streams := []nt.Streams{
		{Name: nt.Orders, Subjects: []nt.Subject{nt.OrderCreated, nt.OrderUpdated}},
		{Name: nt.Expiration, Subjects: []nt.Subject{nt.ExpirationComplete}},
	}

	return &Server{
		nats:  nt.GetNats(subs, streams, "expiration-service"),
		async: as.GetAsync(),
		echo:  getEcho(),
	}
}

func (srv *Server) run() {
	srv.start()
	srv.waitForSignals()
	srv.shutdown()
}

func (srv *Server) start() {
	srv.async.Start(&srv.wg)
	srv.nats.Start(&srv.wg)
	srv.echo.start(&srv.wg)
}

func (srv *Server) waitForSignals() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGTSTP)

	for {
		<-sigs
		break
	}
}

func (srv *Server) shutdown() {
	srv.nats.Shutdown()
	// async also listens the same signals and stops by itself
	srv.echo.shutdown()
	srv.wg.Wait()
	log.Info("Server stopped")
}

func main() {
	srv := newServer()
	srv.run()
}

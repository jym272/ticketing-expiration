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
	logger := log.New()
	return &Server{
		nats:  nt.GetNats(subs),
		async: as.GetAsync(),
		echo:  newEcho(logger),
	}
}

func (srv *Server) Run() error {
	if err := srv.start(); err != nil {
		return err
	}
	srv.waitForSignals()
	srv.shutdown()
	return nil
}

func (srv *Server) start() error {

	srv.async.Start(&srv.wg)
	srv.nats.Start(&srv.wg)
	srv.echo.start(&srv.wg)
	// EL SHUTDOWN usa signals

	return nil
}

func (srv *Server) shutdown() {
	//srv.nats.Stop()
	//srv.async.Stop()
	srv.echo.shutdown()
	srv.wg.Wait()

}

func main() {

	srv := NewServer()
	if err := srv.Run(); err != nil {
		log.Fatalf("could not run server: %v", err)
	}

}

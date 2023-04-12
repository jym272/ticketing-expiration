package async

import (
	"fmt"
	"os"
	"sync"
	"time"
	"workspace/nats"

	log "github.com/sirupsen/logrus"

	"github.com/hibiken/asynq"
)

type Priority int

const (
	concurrentWorkers   = 10
	criticalPriority    = Priority(6)
	defaultPriority     = Priority(3)
	lowPriority         = Priority(1)
	healthCheckInterval = 2 * time.Second
)

func redisHealthFunc(err error) {
	if err != nil {
		log.Panic("a redis error has been found: ", err)
	}
}

type Async struct {
	srv *asynq.Server
	mux *asynq.ServeMux
}

func (a *Async) Start(wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := a.srv.Run(a.mux); err != nil {
			log.Fatalf("could not run the async server: %v", err)
		}
	}()
}

func GetAsync() (srv *Async) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" || redisPort == "" {
		log.Panic("REDIS_HOST or REDIS_PORT is not set")
	}

	url := fmt.Sprintf("%s:%s", redisHost, redisPort)

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: url},
		asynq.Config{
			Concurrency: concurrentWorkers,
			Queues: map[string]int{
				"critical": int(criticalPriority),
				"default":  int(defaultPriority),
				"low":      int(lowPriority),
			},
			Logger:              log.New(),
			HealthCheckInterval: healthCheckInterval,
			HealthCheckFunc:     redisHealthFunc,
		},
	)

	mux := asynq.NewServeMux()
	mux.Handle(string(nats.OrderCreated), NewProcessor(nats.OrderCreated))

	return &Async{
		srv: server,
		mux: mux,
	}
}

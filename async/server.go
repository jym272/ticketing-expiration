package async

import (
	"os"
	"workspace/nats"

	log "github.com/sirupsen/logrus"

	"github.com/hibiken/asynq"
)

type Priority int

const (
	concurrentWorkers = 10
	redisAddr         = "127.0.0.1:7157"
	criticalPriority  = Priority(6)
	defaultPriority   = Priority(3)
	lowPriority       = Priority(1)
)

func StartServer() {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = redisAddr
	}

	l := log.WithFields(log.Fields{
		"redis_url": url,
	})

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: url},
		asynq.Config{
			Concurrency: concurrentWorkers,
			Queues: map[string]int{
				"critical": int(criticalPriority),
				"default":  int(defaultPriority),
				"low":      int(lowPriority),
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.Handle(string(nats.OrderCreated), NewOrderProcessor(nats.OrderCreated))

	if err := srv.Run(mux); err != nil {
		l.Fatalf("could not run server: %v", err)
	}
}

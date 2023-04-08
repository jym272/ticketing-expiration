package async1

import (
	"github.com/hibiken/asynq"
	"log"
	"workspace/async1/tasks"
	"workspace/nats"
)

const redisAddr = "127.0.0.1:7157"

func StartAsyncServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	mux := asynq.NewServeMux()
	mux.Handle(nats.OrderCreated, tasks.NewOrderProcessor(nats.OrderCreated))

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

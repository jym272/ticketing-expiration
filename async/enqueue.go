package async

import (
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

const maxRetry = 10
const timeout = 10 * time.Second

func EnqueueTask(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" || redisPort == "" {
		log.Panic("REDIS_HOST or REDIS_PORT is not set")
	}

	url := fmt.Sprintf("%s:%s", redisHost, redisPort)

	l := log.WithFields(log.Fields{
		"redis_url": url,
	})

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: url})
	defer func(client *asynq.Client) {
		err := client.Close()
		if err != nil {
			l.Fatalf("could not close client: %v", err)
		}
	}(client)

	opts = append(opts, asynq.MaxRetry(maxRetry), asynq.Timeout(timeout))
	info, err := client.Enqueue(task, opts...)

	if err != nil {
		l.Errorf("could not enqueue task: %v", err)
		return nil, err
	}

	return info, nil
}

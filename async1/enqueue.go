package async1

import (
	"github.com/hibiken/asynq"
	"log"
	"os"
	"time"
)

func EnqueueOrder(task *asynq.Task, expiresAt time.Time) (*asynq.TaskInfo, error) {

	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = redisAddr
	}
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: url})
	defer func(client *asynq.Client) {
		err := client.Close()
		if err != nil {
			log.Fatalf("could not close client: %v", err)
		}
	}(client)

	// TODO: better ioptions to make sure that the task was complete
	info, err := client.Enqueue(task, asynq.ProcessAt(expiresAt), asynq.MaxRetry(10), asynq.Timeout(10*time.Second))
	if err != nil {
		log.Fatalf("could not schedule task: %v", err)
		return nil, err
	}
	return info, nil
}

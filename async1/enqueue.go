package async1

import (
	"fmt"
	"github.com/hibiken/asynq"
	"log"
	"os"
	"time"
)

func EnqueueOrder(task *asynq.Task, expiresAt string) (*asynq.TaskInfo, error) {

	//expiresAt string is in UTC iso8601 format
	duration, err := time.Parse(time.RFC3339, expiresAt)
	fmt.Println("duration:", duration)
	if err != nil {
		log.Fatalf("could not parse time: %v", err)
		return nil, err
	}
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
	info, err := client.Enqueue(task, asynq.ProcessAt(duration))
	if err != nil {
		log.Fatalf("could not schedule task: %v", err)
		return nil, err
	}
	return info, nil
}

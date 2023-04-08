package async1

import (
	"github.com/hibiken/asynq"
	"log"
	"time"
)

func EnqueueOrder(task *asynq.Task, expiresAt string) (*asynq.TaskInfo, error) {

	//expiresAt to time
	duration, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		log.Fatalf("could not parse time: %v", err)
		return nil, err
	}
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
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

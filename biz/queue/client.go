package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

var client *asynq.Client

func InitClient(redisAddr, redisPassword string, redisDB int) {
	client = asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
}

func EnqueueReview(taskType string, payload ReviewPayload) error {
	if client == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(taskType, data)
	_, err = client.Enqueue(task)
	return err
}

func CloseClient() {
	if client != nil {
		client.Close()
	}
}

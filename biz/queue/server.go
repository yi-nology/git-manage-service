package queue

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	codereview "github.com/yi-nology/git-manage-service/biz/service/codereview"
)

func NewServer(redisAddr, redisPassword string, redisDB int) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskReviewFull, handleReviewFull)
	mux.HandleFunc(TaskReviewRetry, handleReviewFull)
	mux.HandleFunc(TaskReviewLLM, handleReviewFull)
	return mux
}

func StartWorker(redisAddr, redisPassword string, redisDB int) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword, DB: redisDB},
		asynq.Config{Concurrency: 5},
	)

	mux := NewServer(redisAddr, redisPassword, redisDB)
	log.Println("[Asynq] Worker starting...")
	if err := srv.Run(mux); err != nil {
		log.Printf("[Asynq] Worker stopped: %v", err)
	}
}

func handleReviewFull(ctx context.Context, t *asynq.Task) error {
	var p ReviewPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	return codereview.RunReview(ctx, p.TaskID)
}

package queue

const (
	TaskReviewFull  = "review:full"
	TaskReviewRetry = "review:retry"
	TaskReviewLLM   = "review:llm"
)

type ReviewPayload struct {
	TaskID uint `json:"task_id"`
}

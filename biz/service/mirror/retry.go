package mirror

import (
	"math"
	"time"
)

type RetryStrategy struct {
	MaxRetry int
}

func NewRetryStrategy(maxRetry int) *RetryStrategy {
	if maxRetry <= 0 {
		maxRetry = 5
	}
	return &RetryStrategy{MaxRetry: maxRetry}
}

func (r *RetryStrategy) ShouldRetry(retryCount int) bool {
	return retryCount < r.MaxRetry
}

func (r *RetryStrategy) GetNextRetryDelay(retryCount int) time.Duration {
	switch retryCount {
	case 1:
		return 1 * time.Minute
	case 2:
		return 5 * time.Minute
	case 3:
		return 30 * time.Minute
	case 4:
		return 2 * time.Hour
	default:
		return time.Duration(math.Pow(2, float64(retryCount))) * time.Hour
	}
}

func (r *RetryStrategy) GetNextSyncAt(retryCount int) time.Time {
	return time.Now().Add(r.GetNextRetryDelay(retryCount))
}

func (r *RetryStrategy) ShouldPause(retryCount int) bool {
	return retryCount >= r.MaxRetry
}

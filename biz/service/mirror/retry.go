package mirror

import (
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
		// 小时级指数阶梯；retryCount≥8 时已超 7 天封顶，直接返回。
		// 不能只 clamp 移位位数——(1<<30)*time.Hour 仍会溢出为负的 Duration。
		const maxDelay = 168 * time.Hour
		if retryCount >= 8 {
			return maxDelay
		}
		return (1 << uint(retryCount)) * time.Hour
	}
}

func (r *RetryStrategy) GetNextSyncAt(retryCount int) time.Time {
	return time.Now().Add(r.GetNextRetryDelay(retryCount))
}

func (r *RetryStrategy) ShouldPause(retryCount int) bool {
	return retryCount >= r.MaxRetry
}

package worker

type RetryStrategy struct {
	maxRetries int
}

func NewRetryStrategy(maxRetries int) *RetryStrategy {
	if maxRetries < 0 {
		panic("max retries cannot be negative")
	}

	return &RetryStrategy{
		maxRetries: maxRetries,
	}
}

func (r *RetryStrategy) ShouldRetry(retryCount int) bool {
	return retryCount < r.maxRetries
}

func (r *RetryStrategy) MaxRetries() int {
	return r.maxRetries
}

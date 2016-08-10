package ratelimiter

import "golang.org/x/time/rate"

type limiter struct {
	rateLimiter *rate.Limiter
	quota       Quota
}

func (l *limiter) allow() bool {
	return l.rateLimiter.Allow()
}

func (l *limiter) rpm() string {
	return l.quota.rpm()
}

func (l *limiter) retryAfter() string {
	return l.quota.retryAfter()
}

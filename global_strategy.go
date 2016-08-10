package ratelimiter

import (
	"net/http"

	"golang.org/x/time/rate"
)

type globalStrategy struct {
	opts    opts
	limiter *limiter
}

func newGlobalStrategy(opts opts) *globalStrategy {
	quota := opts.getQuota("")
	return &globalStrategy{limiter: &limiter{quota: quota, rateLimiter: rate.NewLimiter(quota.limit(), quota.reqs)}}
}
func (g *globalStrategy) getLimiter(r *http.Request) *limiter {
	return g.limiter
}

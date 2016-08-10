package ratelimiter

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type keyStrategy struct {
	*sync.Mutex
	opts     opts
	limiters map[string]*limiter
}

func newKeysStrategy(opts opts) *keyStrategy {
	return &keyStrategy{&sync.Mutex{}, opts, map[string]*limiter{}}
}
func (g *keyStrategy) getLimiter(r *http.Request) *limiter {
	g.Lock()
	defer g.Unlock()

	key := g.opts.getKey(r)
	lim, found := g.limiters[key]
	if !found {
		quota := g.opts.getQuota(key)
		rateLimiter := rate.NewLimiter(quota.limit(), quota.reqs)
		lim = &limiter{rateLimiter: rateLimiter, quota: quota}
		g.limiters[key] = lim
	}
	return lim
}

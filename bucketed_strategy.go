package ratelimiter

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type entry struct {
	*sync.Mutex
	limiters map[string]*limiter
}
type bucketedStrategy struct {
	opts    opts
	buckets []*entry
}

func newBucketsAndKeysStrategy(opts opts) *bucketedStrategy {
	buckets := make([]*entry, opts.buckets)
	for i := 0; i < opts.buckets; i++ {
		buckets[i] = &entry{&sync.Mutex{}, map[string]*limiter{}}
	}
	return &bucketedStrategy{opts: opts, buckets: buckets}
}
func (g *bucketedStrategy) getLimiter(r *http.Request) *limiter {
	key := g.opts.getKey(r)
	hasher := g.opts.getHasher()
	hasher.Write([]byte(key))
	h := hasher.Sum32()
	e := g.buckets[h%uint32(g.opts.buckets)]

	e.Lock()
	defer e.Unlock()

	lim, found := e.limiters[key]
	if !found {
		quota := g.opts.getQuota(key)
		rateLimiter := rate.NewLimiter(quota.limit(), quota.reqs)
		lim = &limiter{rateLimiter: rateLimiter, quota: quota}
		e.limiters[key] = lim
	}
	return lim
}

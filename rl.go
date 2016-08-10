package ratelimiter

import (
	"net/http"

	"github.com/urfave/negroni"
)

type limitStrategy interface {
	getLimiter(*http.Request) *limiter
}

type rl struct {
	opts     opts
	strategy limitStrategy
}

func newRL(opts opts) negroni.Handler {
	l := &rl{opts: opts}
	// Decide which limiter func we are going to use depending on the set options
	if opts.getKey != nil {
		if opts.buckets > 0 && opts.getHasher != nil {
			// We want to use buckets and keys limiter
			l.strategy = newBucketsAndKeysStrategy(opts)
		} else {
			// We want to use only keys limiter
			l.strategy = newKeysStrategy(opts)
		}
	} else {
		// We want a global limiter
		l.strategy = newGlobalStrategy(opts)
	}

	return l
}

func (m *rl) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	limiter := m.strategy.getLimiter(req)

	ok := limiter.allow()

	h := rw.Header()
	h.Set("X-RateLimit-Limit", limiter.rpm())

	if ok {
		next(rw, req)
	} else {
		h.Set("Retry-After", limiter.retryAfter())
		rw.WriteHeader(m.opts.status)
	}
}

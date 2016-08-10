package ratelimiter

import (
	"hash"
	"net/http"
	"time"

	"github.com/urfave/negroni"
)

type GetKeyFunc func(*http.Request) string
type GetQuotaFunc func(string) Quota
type GetHasherFunc func() hash.Hash32

type opts struct {
	status    int
	buckets   int
	getKey    GetKeyFunc
	getHasher GetHasherFunc
	getQuota  GetQuotaFunc
}

func getDefaults() opts {
	return opts{
		status:    http.StatusTooManyRequests,
		buckets:   0,
		getKey:    nil,
		getHasher: nil,
		getQuota: func(key string) Quota {
			return InfQuota
		},
	}
}

func NewGlobal() opts {
	op := getDefaults()
	return op
}
func NewLimitByKeys(getKey GetKeyFunc) opts {
	op := getDefaults()
	op.getKey = getKey
	op.getHasher = nil
	op.buckets = 0
	return op
}
func NewLimitByBucketedKeys(buckets int, getHasher GetHasherFunc, getKey GetKeyFunc) opts {
	op := getDefaults()
	op.getKey = getKey
	op.getHasher = getHasher
	op.buckets = buckets

	return op
}

func (o opts) WithQuotaByKeys(getQuota GetQuotaFunc) opts {
	o.getQuota = getQuota
	return o
}
func (o opts) WithStatus(status int) opts {
	o.status = status
	return o
}
func (o opts) WithDefaultQuota(reqs int, interval time.Duration) opts {
	o.getQuota = func(key string) Quota {
		return NewQuota(reqs, interval)
	}
	return o
}
func (o opts) Middleware() negroni.Handler {
	return newRL(o)
}

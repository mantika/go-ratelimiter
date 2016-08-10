package ratelimiter

import (
	"math"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

var InfQuota Quota = Quota{
	reqs:        math.MaxInt32,
	interval:    rate.InfDuration,
	_limit:      rate.Inf,
	_rpm:        "unlimited",
	_retryAfter: "",
}

type Quota struct {
	reqs        int
	interval    time.Duration
	_rpm        string
	_limit      rate.Limit
	_retryAfter string
}

func (q Quota) limit() rate.Limit {
	return q._limit
}

func (q Quota) rpm() string {
	return q._rpm
}
func (q Quota) retryAfter() string {
	return q._retryAfter
}

func NewQuota(reqs int, interval time.Duration) Quota {
	return Quota{
		reqs:        reqs,
		interval:    interval,
		_limit:      rate.Limit(float64(reqs) / interval.Seconds()),
		_rpm:        strconv.Itoa(int(math.Floor(float64(reqs) / interval.Minutes()))),
		_retryAfter: strconv.FormatInt(int64(interval.Seconds()), 10),
	}
}

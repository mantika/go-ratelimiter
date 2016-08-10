package ratelimiter

import (
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/stretchr/testify/assert"
)

func TestQuotaLimit(t *testing.T) {
	assert.Equal(t, rate.Limit(1), NewQuota(60, time.Minute).limit())
	assert.Equal(t, rate.Limit(2), NewQuota(2, time.Second).limit())
}
func TestQuotaRPM(t *testing.T) {
	assert.Equal(t, "60", NewQuota(1, time.Second).rpm())
}
func TestQuotaRetryAfter(t *testing.T) {
	assert.Equal(t, "2", NewQuota(1, 2*time.Second).retryAfter())
}

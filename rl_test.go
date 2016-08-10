package ratelimiter

import (
	"fmt"
	"hash"
	"hash/fnv"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/negroni"
)

func TestDefaultStatus(t *testing.T) {
	m := NewGlobal().
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	resp1, err1 := http.Get(ts.URL)
	resp2, err2 := http.Get(ts.URL)

	assert.Nil(t, err1)
	assert.Equal(t, 200, resp1.StatusCode)

	assert.Nil(t, err2)
	assert.Equal(t, 429, resp2.StatusCode)
}

func TestCustomStatus(t *testing.T) {
	m := NewGlobal().
		WithDefaultQuota(0, time.Second).
		WithStatus(http.StatusBadRequest).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	resp1, err1 := http.Get(ts.URL)

	assert.Nil(t, err1)
	assert.Equal(t, 400, resp1.StatusCode)
}

func TestHeaders(t *testing.T) {
	m := NewGlobal().
		WithDefaultQuota(1, 2*time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	resp1, _ := http.Get(ts.URL)
	resp2, _ := http.Get(ts.URL)

	assert.Equal(t, 200, resp1.StatusCode)
	assert.Equal(t, 429, resp2.StatusCode)

	// X-RateLimit-Limit is expected in RPM
	assert.Equal(t, "30", resp1.Header.Get("X-RateLimit-Limit"))
	// There should be no Retry-After header if request succeded
	assert.Equal(t, "", resp1.Header.Get("Retry-After"))

	assert.Equal(t, "30", resp2.Header.Get("X-RateLimit-Limit"))
	// Retry-After should be represented in seconds with no decimals
	assert.Equal(t, "2", resp2.Header.Get("Retry-After"))
}
func TestRateByKey(t *testing.T) {
	getKey := func(req *http.Request) string {
		return req.URL.Query().Get("account_id")
	}

	m := NewLimitByKeys(getKey).
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	resp1, _ := http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
	resp2, _ := http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
	resp3, _ := http.Get(fmt.Sprintf("%s?account_id=2", ts.URL))

	assert.Equal(t, 200, resp1.StatusCode)
	assert.Equal(t, 429, resp2.StatusCode)
	assert.Equal(t, 200, resp3.StatusCode)
}
func TestRateByKeyWithDifferentQuotas(t *testing.T) {
	getKey := func(req *http.Request) string {
		return req.URL.Query().Get("account_id")
	}
	getQuota := func(key string) Quota {
		if key == "2" {
			return NewQuota(1, time.Minute)
		}
		return NewQuota(2, time.Second)
	}

	m := NewLimitByKeys(getKey).
		WithQuotaByKeys(getQuota).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	r, _ := http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
	assert.Equal(t, 200, r.StatusCode)
	r, _ = http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
	assert.Equal(t, 200, r.StatusCode)
	r, _ = http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
	assert.Equal(t, 429, r.StatusCode)

	r, _ = http.Get(fmt.Sprintf("%s?account_id=2", ts.URL))
	assert.Equal(t, 200, r.StatusCode)
	r, _ = http.Get(fmt.Sprintf("%s?account_id=2", ts.URL))
	assert.Equal(t, 429, r.StatusCode)
}

func TestRateByKeyRaceCondition(t *testing.T) {
	getKey := func(req *http.Request) string {
		return req.URL.Query().Get("account_id")
	}

	m := NewLimitByKeys(getKey).
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for i := 0; i < 10; i++ {
			http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 10; i++ {
			http.Get(fmt.Sprintf("%s?account_id=%d", ts.URL, i))
		}
		wg.Done()
	}()
	wg.Wait()
}
func TestRateByBucketedKey(t *testing.T) {
	getKey := func(req *http.Request) string {
		return req.URL.Query().Get("account_id")
	}
	getHash := func() hash.Hash32 {
		return fnv.New32a()
	}

	m := NewLimitByBucketedKeys(10, getHash, getKey).
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	resp1, _ := http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
	resp2, _ := http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
	resp3, _ := http.Get(fmt.Sprintf("%s?account_id=2", ts.URL))

	assert.Equal(t, 200, resp1.StatusCode)
	assert.Equal(t, 429, resp2.StatusCode)
	assert.Equal(t, 200, resp3.StatusCode)
}

func TestRateByBucketedKeyRaceCondition(t *testing.T) {
	getKey := func(req *http.Request) string {
		return req.URL.Query().Get("account_id")
	}
	getHash := func() hash.Hash32 {
		return fnv.New32a()
	}

	m := NewLimitByBucketedKeys(10, getHash, getKey).
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for i := 0; i < 10; i++ {
			http.Get(fmt.Sprintf("%s?account_id=1", ts.URL))
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i < 10; i++ {
			http.Get(fmt.Sprintf("%s?account_id=%d", ts.URL, i))
		}
		wg.Done()
	}()
	wg.Wait()
}

func BenchmarkWithout(b *testing.B) {
	n := negroni.New()
	ts := httptest.NewServer(n)
	defer ts.Close()

	b.N = 1000
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			http.Get(ts.URL)
		}
	})
}
func BenchmarkGlobal(b *testing.B) {
	m := NewGlobal().
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	b.N = 1000
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			http.Get(ts.URL)
		}
	})
}
func BenchmarkKeys(b *testing.B) {
	getKey := func(req *http.Request) string {
		return req.URL.Query().Get("account_id")
	}

	m := NewLimitByKeys(getKey).
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)
	ts := httptest.NewServer(n)
	defer ts.Close()

	b.N = 1000
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			http.Get(fmt.Sprintf("%s?account_id=%d", ts.URL, i))
			i++
		}
	})
}
func BenchmarkBucketedKeys(b *testing.B) {
	getKey := func(req *http.Request) string {
		return req.URL.Query().Get("account_id")
	}
	getHash := func() hash.Hash32 {
		return fnv.New32a()
	}

	m := NewLimitByBucketedKeys(100, getHash, getKey).
		WithDefaultQuota(1, time.Second).
		Middleware()
	n := negroni.New(m)

	ts := httptest.NewServer(n)
	defer ts.Close()

	b.N = 1000
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			http.Get(fmt.Sprintf("%s?account_id=%d", ts.URL, i))
			i++
		}
	})
}

package middleware

import (
	"sync"
	"time"
)

type Limiter struct {
	mu            sync.Mutex
	window        time.Duration
	bucketSize    time.Duration
	max           int
	buckets       []int
	total         int
	currentBucket int
	lastTime      time.Time
	now           func() time.Time
}

func New(max int, window, bucketSize time.Duration) *Limiter {
	if max < 0 {
		panic("max must be >= 0")
	}
	if window <= 0 || bucketSize <= 0 {
		panic("durations must be > 0")
	}
	if window%bucketSize != 0 {
		panic("window must be multiple of bucketSize")
	}
	numBuckets := int(window / bucketSize)
	now := time.Now()
	last := now.Truncate(bucketSize)
	cb := int(last.UnixNano()/bucketSize.Nanoseconds()) % numBuckets
	if cb < 0 {
		cb += numBuckets
	}
	return &Limiter{
		window:        window,
		bucketSize:    bucketSize,
		max:           max,
		buckets:       make([]int, numBuckets),
		total:         0,
		currentBucket: cb,
		lastTime:      last,
		now:           time.Now,
	}
}

func (l *Limiter) advance(now time.Time) {
	nowBucket := now.Truncate(l.bucketSize)
	if nowBucket.Before(l.lastTime) {
		l.resetLocked(nowBucket)
		return
	}
	passed := int(nowBucket.Sub(l.lastTime) / l.bucketSize)
	if passed == 0 {
		return
	}
	if passed >= len(l.buckets) {
		l.resetLocked(nowBucket)
		return
	}
	for i := 0; i < passed; i++ {
		l.currentBucket = (l.currentBucket + 1) % len(l.buckets)
		l.total -= l.buckets[l.currentBucket]
		l.buckets[l.currentBucket] = 0
	}
	if l.total < 0 {
		l.total = 0
	}
	l.lastTime = nowBucket
}

func (l *Limiter) resetLocked(nowBucket time.Time) {
	for i := range l.buckets {
		l.buckets[i] = 0
	}
	l.total = 0
	l.lastTime = nowBucket
	if len(l.buckets) > 0 {
		l.currentBucket = int(nowBucket.UnixNano()/l.bucketSize.Nanoseconds()) % len(l.buckets)
		if l.currentBucket < 0 {
			l.currentBucket += len(l.buckets)
		}
	} else {
		l.currentBucket = 0
	}
}

func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.resetLocked(l.now().Truncate(l.bucketSize))
}

func (l *Limiter) Allow() bool {
	return l.AllowN(1)
}

func (l *Limiter) AllowN(n int) bool {
	if n <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.advance(now)
	if l.total+n <= l.max {
		l.buckets[l.currentBucket] += n
		l.total += n
		return true
	}
	return false
}

func (l *Limiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.advance(now)
	rem := l.max - l.total
	if rem < 0 {
		return 0
	}
	return rem
}

func (l *Limiter) SetMax(max int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.max = max
}

func (l *Limiter) Snapshot() (total int, buckets []int, last time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cpy := make([]int, len(l.buckets))
	copy(cpy, l.buckets)
	return l.total, cpy, l.lastTime
}

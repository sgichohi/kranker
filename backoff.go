package main

import (
	"math"
	"math/rand"
	"time"
)

// BackoffPolicy implements a backoff policy, randomizing its delays
// and saturating at the final value in Millis.
type BackoffPolicy struct {
	Millis []int
}

var defaultBackoff BackoffPolicy

func init() {
	defaultBackoff = makeBackoffPolicy(MAXATTEMPTS)
}

func makeBackoffPolicy(length int) BackoffPolicy {

	rand.Seed(5)
	b := make([]int, length)
	b[0] = 0
	r := rand.Intn(length)
	for i := 1; i < length; i++ {
		b[i] = int(math.Pow(float64(i), float64(r)))
	}

	return BackoffPolicy{Millis: b}

}

// Duration returns the time duration of the n'th wait cycle in a
// backoff policy. This is b.Millis[n], randomized to avoid thundering
// herds.
func (b BackoffPolicy) Duration(n int) time.Duration {
	if n >= len(b.Millis) {
		n = len(b.Millis) - 1
	}

	return time.Duration(jitter(b.Millis[n])) * time.Millisecond
}

// jitter returns a random integer uniformly distributed in the range
// [0.5 * millis .. 1.5 * millis]
func jitter(millis int) int {
	if millis == 0 {
		return 0
	}

	return millis/2 + rand.Intn(millis)
}

package routing

import (
	"context"
	"math"
	"time"
)

type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

func (p RetryPolicy) Delay(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}
	d := float64(p.InitialDelay) * math.Pow(2, float64(attempt-1))
	if d > float64(p.MaxDelay) {
		d = float64(p.MaxDelay)
	}
	return time.Duration(d)
}
func RunWithRetry(ctx context.Context, p RetryPolicy, fn func(context.Context) error) error {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}
	for i := 1; i <= p.MaxAttempts; i++ {
		if err := fn(ctx); err == nil {
			return nil
		}
		if i == p.MaxAttempts {
			break
		}
		timer := time.NewTimer(p.Delay(i))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return context.Canceled
}

type Circuit struct {
	Failures  int
	OpenUntil time.Time
	Threshold int
	Cooldown  time.Duration
}

func (c *Circuit) Allow(now time.Time) bool { return c.OpenUntil.IsZero() || now.After(c.OpenUntil) }
func (c *Circuit) Record(success bool, now time.Time) {
	if success {
		c.Failures = 0
		c.OpenUntil = time.Time{}
		return
	}
	c.Failures++
	if c.Failures >= c.Threshold {
		c.OpenUntil = now.Add(c.Cooldown)
	}
}

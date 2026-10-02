package worker

import "time"

type BackoffFunc func(attempt int, baseSeconds int) time.Duration;

func FixedBackoff(attempt int, baseSeconds int) time.Duration {
	return time.Duration(baseSeconds) * time.Second;
}
package dispatcher

import (
	"errors"
)

type Job struct {
	ID string;
	Payload string;
}

var ErrEmpty = errors.New("queue is empty");

type Dispatcher interface {
	Enqueue(payload string) (string, error);
	Dequeue() (*Job, error);
	Ack(jobId string) error;
}
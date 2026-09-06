package dispatcher

import (
	"errors"

	"github.com/simritkaul/queueforge/queue"
)

type InProcessDispatcher struct {
	qf *queue.Queue;
}

var _ Dispatcher = (*InProcessDispatcher)(nil)

func NewInprocessDispatcher(qf *queue.Queue) *InProcessDispatcher {
	return &InProcessDispatcher{
		qf: qf,
	}
}

func (ipd *InProcessDispatcher) Enqueue(payload string) (string, error) {
	return ipd.qf.Enqueue(payload);
}

func (ipd *InProcessDispatcher) Dequeue() (*Job, error) {
	queueJob, err := ipd.qf.Dequeue();
	if errors.Is(err, queue.ErrEmpty) {
		return nil, ErrEmpty;
	} else if err != nil {
		return nil, err;
	}

	job := &Job{
		ID : queueJob.ID,
		Payload: queueJob.Payload,
	}

	return job, nil;
}

func (ipd *InProcessDispatcher) Ack(jobId string) error {
	return ipd.qf.Ack(jobId);
}
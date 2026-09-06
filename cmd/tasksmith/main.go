package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/simritkaul/queueforge/queue"
	"github.com/simritkaul/tasksmith/internal/dispatcher"
)

func main() {
	wal, err := queue.OpenWAL("tasksmith.wal")
	if err != nil {
		log.Fatal(err)
	}
	defer wal.Close()

	fmt.Println("WAL opened successfully")

	q, err := queue.NewQueue(wal)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Queue constructed successfully")

	disp := dispatcher.NewInprocessDispatcher(q);

	// Enqueue
	jobId, err := disp.Enqueue("Hello World")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Enqueued Job: ", jobId)

	// Dequeue
	job, err := disp.Dequeue()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Dequeued Job:  %+v\n", job)

	// Ack
	if err := disp.Ack(job.ID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Acked Job: ", job.ID)

	// Dequeue again (should be empty)
	_, err = disp.Dequeue()
	if errors.Is(err, dispatcher.ErrEmpty) {
		fmt.Println("Queue is empty as expected")
	} else if err != nil {
		log.Fatal(err)
	} else {
		log.Fatal("expected empty queue but found a job")
	}

}


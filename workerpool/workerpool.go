package workerpool

import (
	"log"
	"time"
)

type Job struct {
	Name     string
	Duration time.Duration
	Done     chan struct{} // Notify handler when done
}

func StartWorkerPool(workerCount, queueSize int) chan Job {
	jobQueue := make(chan Job, queueSize)

	for i := 1; i <= workerCount; i++ {
		go func(id int, jobs <-chan Job) { //Read-only channel using <-chan Job instead of chan Job
			for job := range jobs {
				log.Printf("Worker %d: Starting '%s' for %v\n", id, job.Name, job.Duration)
				time.Sleep(job.Duration)
				log.Printf("Worker %d: Finished '%s'\n", id, job.Name)
				close(job.Done)
			}
		}(i, jobQueue)
	}

	return jobQueue
}

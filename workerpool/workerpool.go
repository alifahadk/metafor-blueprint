package workerpool

import (
	"fmt"
	"log"
	"rimcs/metafor-blueprint/client"
	"time"
)

/* apis: [

     'hello' : {
          processing rate (duration)
          downstream_calls : [


              'world': {
                 'type':  blocking, 'url', 'world.com', 'order' '1'

              },

               'I am machine': {
                'type':  non-blocking, 'url', 'machine.com', 'order' '1'

             }

          ]
     }

   ]

*/

type DownstreamCall struct {
	Type     bool
	URL      string
	Endpoint string
	Retry    bool
}

type Job struct {
	Name            string
	Duration        time.Duration
	Done            chan struct{} // Notify handler when done
	DownstreamCalls []DownstreamCall
}

func StartWorkerPool(workerCount, queueSize int) chan Job {
	jobQueue := make(chan Job, queueSize)

	for i := 1; i <= workerCount; i++ {
		go func(id int, jobs <-chan Job) { //Read-only channel using <-chan Job instead of chan Job
			for job := range jobs {
				log.Printf("Worker %d: Starting '%s' for %v\n", id, job.Name, job.Duration)
				for _, d := range job.DownstreamCalls {
					t, _ := time.ParseDuration("1s")
					_, _, _, err := client.GetWithRetry(d.URL, d.Endpoint, t, 3)
					if err != nil {
						fmt.Println("downstream call error:", err)
					}
				}
				time.Sleep(job.Duration)
				log.Printf("Worker %d: Finished '%s'\n", id, job.Name)
				close(job.Done)
			}
		}(i, jobQueue)
	}

	return jobQueue
}

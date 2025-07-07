package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func main() {
	var (
		rps      = flag.Int("rps", 5, "Requests per second")
		duration = flag.String("duration", "2s", "Job duration (e.g., 500ms, 2s)")
		total    = flag.Int("total", 20, "Total number of requests to send")
	)
	flag.Parse()

	client := &http.Client{Timeout: 10 * time.Second}
	interval := time.Second / time.Duration(*rps)

	fmt.Printf("Sending %d requests at %d req/sec with job duration: %s\n", *total, *rps, *duration)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sent := 0
	for range ticker.C {
		if sent >= *total {
			break
		}
		go func(id int) {
			workParam := fmt.Sprintf("test%d:%s", id, *duration)
			endpoint := "http://localhost:8080/workload?work=" + url.QueryEscape(workParam)

			resp, err := client.Get(endpoint)
			if err != nil {
				fmt.Printf("[ERROR] Request %d failed: %v\n", id, err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("[OK] Request %d got status: %s\n", id, resp.Status)
		}(sent)
		sent++
	}

	// Wait for last goroutines to finish
	time.Sleep(2 * time.Second)
}

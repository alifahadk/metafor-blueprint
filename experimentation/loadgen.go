package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	var (
		normalRPS       = flag.Int("normalRPS", 5, "Normal requests per second")
		triggerRPS      = flag.Int("triggerRPS", 15, "Triggered requests per second")
		timeout         = flag.String("timeout", "5s", "Request timeout")
		duration        = flag.String("duration", "12s", "Total test duration")
		triggerDuration = flag.String("triggerDuration", "4s", "Duration of trigger load in the middle")
		outputFile      = flag.String("out", "results.csv", "CSV file to write results to")
		maxRetries      = flag.Int("retries", 3, "Maximum retries per request")
	)
	flag.Parse()

	// Parse durations
	totalDur, err := time.ParseDuration(*duration)
	if err != nil {
		panic("Invalid -duration: " + err.Error())
	}
	triggerDur, err := time.ParseDuration(*triggerDuration)
	if err != nil {
		panic("Invalid -triggerDuration: " + err.Error())
	}
	timeoutDur, err := time.ParseDuration(*timeout)
	if err != nil {
		panic("Invalid -timeout: " + err.Error())
	}

	// Calculate trigger window
	triggerStart := (totalDur - triggerDur) / 2
	triggerEnd := triggerStart + triggerDur

	client := &http.Client{Timeout: timeoutDur}
	var wg sync.WaitGroup
	startTime := time.Now()
	idCounter := 0

	// Create and open CSV file
	file, err := os.Create(*outputFile)
	if err != nil {
		panic("Failed to open file: " + err.Error())
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"Start", "Duration", "IsError"})

	var mu sync.Mutex // to synchronize writes

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for now := range ticker.C {
		elapsed := time.Since(startTime)
		if elapsed > totalDur {
			break
		}

		// Decide current RPS based on time window
		var currentRPS int
		switch {
		case elapsed < triggerStart:
			currentRPS = *normalRPS
		case elapsed >= triggerStart && elapsed < triggerEnd:
			currentRPS = *triggerRPS
		default:
			currentRPS = *normalRPS
		}

		interval := time.Second / time.Duration(currentRPS)

		wg.Add(1)
		go func(id int, now time.Time) {
			defer wg.Done()

			endpoint := "http://localhost:12345/Insert"

			start := time.Now()

			var err error
			var resp *http.Response
			for attempt := 0; attempt <= *maxRetries; attempt++ {
				resp, err = client.Get(endpoint)
				if err == nil && resp.StatusCode < 400 {
					break
				}
				if resp != nil {
					resp.Body.Close()
				}
				time.Sleep(100 * time.Millisecond) // retry backoff
			}

			end := time.Now()
			duration := end.Sub(start)

			isErr := false
			if err != nil || (resp != nil && resp.StatusCode >= 400) {
				if resp != nil {
					fmt.Println(resp.Status)
				} else {
					fmt.Println(err)
				}
				isErr = true
			} else {
				fmt.Println(resp.Status)
				resp.Body.Close()
			}

			row := []string{
				strconv.FormatInt(start.UnixNano(), 10),
				strconv.FormatInt(duration.Nanoseconds(), 10),
				strconv.FormatBool(isErr),
			}
			mu.Lock()
			_ = writer.Write(row)
			mu.Unlock()
		}(idCounter, now)

		idCounter++
		time.Sleep(interval)
	}

	wg.Wait()
}

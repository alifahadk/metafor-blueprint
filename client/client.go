package client

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GetWithRetry sends a GET request to baseURL + endpoint with the given timeout and retries.
// It returns whether the request was successful, how long it took, the response body, and an error (if any).
func GetWithRetry(baseURL, endpoint string, timeout time.Duration, retries int) (bool, time.Duration, []byte, error) {
	client := &http.Client{Timeout: timeout}
	fullURL := baseURL + endpoint

	start := time.Now()

	for attempt := 0; attempt <= retries; attempt++ {
		req, err := http.NewRequest(http.MethodGet, fullURL, nil)
		if err != nil {
			return false, time.Since(start), nil, fmt.Errorf("creating request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			if attempt < retries {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			return false, time.Since(start), nil, fmt.Errorf("request failed after %d attempts: %w", retries+1, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, time.Since(start), nil, fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode >= 500 && attempt < retries {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return true, time.Since(start), body, nil
	}

	return false, time.Since(start), nil, fmt.Errorf("all %d attempts failed", retries+1)
}

// LogRequestResultCSV appends the request result to a CSV file.
func LogRequestResultCSV(success bool, duration time.Duration, err error, url string) error {
	file, fileErr := os.OpenFile("request_log.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return fileErr
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	status := "success"
	errMsg := ""
	if !success {
		status = "failure"
		errMsg = err.Error()
	}

	record := []string{
		time.Now().Format(time.RFC3339),
		url,
		status,
		duration.String(),
		errMsg,
	}

	return writer.Write(record)
}

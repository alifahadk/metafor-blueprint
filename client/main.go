package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano()) // for jitter

	// Command-line flags
	method := flag.String("method", "GET", "HTTP method to use")
	url := flag.String("url", "", "Target URL (required)")
	body := flag.String("body", "", "Request body (for POST/PUT)")
	headers := headerFlags{}
	flag.Var(&headers, "header", "Optional header in 'Key:Value' format (repeatable)")
	timeout := flag.Duration("timeout", 10*time.Second, "Request timeout")
	retries := flag.Int("retries", 0, "Number of retries on failure")
	backoff := flag.Duration("backoff", 1*time.Second, "Base backoff duration for retries")
	jitter := flag.Duration("jitter", 500*time.Millisecond, "Max jitter added to backoff")

	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "Error: --url is required")
		os.Exit(1)
	}

	client := &http.Client{Timeout: *timeout}
	var bodyReader io.Reader
	if *body != "" {
		bodyReader = bytes.NewBufferString(*body)
	}

	for attempt := 0; attempt <= *retries; attempt++ {
		// rewind body if needed
		var reqBody io.Reader
		if *body != "" {
			reqBody = bytes.NewBufferString(*body) // reset buffer
		}

		req, err := http.NewRequest(strings.ToUpper(*method), *url, reqBody)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(1)
		}

		for k, v := range headers.m {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Request failed (attempt %d/%d): %v\n", attempt+1, *retries+1, err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 500 && attempt < *retries {
				fmt.Fprintf(os.Stderr, "Server error (attempt %d/%d): %s. Retrying...\n", attempt+1, *retries+1, resp.Status)
			} else {
				fmt.Printf("Status: %s\n", resp.Status)
				respBody, _ := io.ReadAll(resp.Body)
				fmt.Printf("Response:\n%s\n", string(respBody))
				return
			}
		}

		if attempt < *retries {
			sleepDuration := exponentialBackoffWithJitter(*backoff, *jitter, attempt)
			fmt.Fprintf(os.Stderr, "Waiting %s before next attempt...\n", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}

	fmt.Fprintf(os.Stderr, "All %d attempt(s) failed.\n", *retries+1)
	os.Exit(1)
}

func exponentialBackoffWithJitter(base, jitter time.Duration, attempt int) time.Duration {
	exp := 1 << attempt
	delay := time.Duration(exp) * base
	jitterOffset := time.Duration(rand.Int63n(int64(jitter)))
	return delay + jitterOffset
}

type headerFlags struct {
	m map[string]string
}

func (h *headerFlags) String() string {
	if h.m == nil {
		return ""
	}
	var pairs []string
	for k, v := range h.m {
		pairs = append(pairs, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(pairs, ", ")
}

func (h *headerFlags) Set(value string) error {
	if h.m == nil {
		h.m = make(map[string]string)
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid header format: %s", value)
	}
	h.m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	return nil
}

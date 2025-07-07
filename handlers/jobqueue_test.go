package handlers

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"rimcs/metafor-blueprint/workerpool"
	"testing"
)

func TestWorkloadHandler_SubmitsJob(t *testing.T) {
	jobQueue := make(chan workerpool.Job, 1)

	handler := &Handler{JobQueue: jobQueue}

	// Simulate a worker that immediately completes the job
	go func() {
		job := <-jobQueue
		close(job.Done) // simulate job completion
	}()

	req := httptest.NewRequest("GET", "/workload?work=test:1ms", nil)
	w := httptest.NewRecorder()

	router := httprouter.New()
	router.GET("/workload", handler.Workload())
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestWorkloadHandler_QueueFull(t *testing.T) {
	jobQueue := make(chan workerpool.Job, 0) // no buffer so it blocks immediately

	handler := &Handler{JobQueue: jobQueue}

	req := httptest.NewRequest("GET", "/workload?work=test:1s", nil)
	w := httptest.NewRecorder()

	router := httprouter.New()
	router.GET("/workload", handler.Workload())
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable, got %d", resp.StatusCode)
	}
}

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Workload_Success(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/workload?work=job1:10ms", nil)
	res := httptest.NewRecorder()

	h := handler.Workload()
	h(res, req, nil)

	if res.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", res.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Response body is not valid JSON: %v", err)
	}

	if resp["result"] != "success" {
		t.Errorf("Expected result 'success', got %q", resp["result"])
	}
}

func TestHandler_Workload_MissingWorkloadParam(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/workload", nil)
	res := httptest.NewRecorder()

	h := handler.Workload()
	h(res, req, nil)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	var resp map[string]string
	json.Unmarshal(res.Body.Bytes(), &resp)
	if resp["result"] != "Missing workload parameter" {
		t.Errorf("Expected error message 'Missing workload parameter', got %q", resp["result"])
	}
}

func TestHandler_Workload_InvalidFormat(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/workload?work=invalidformat", nil)
	res := httptest.NewRecorder()

	h := handler.Workload()
	h(res, req, nil)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	var resp map[string]string
	json.Unmarshal(res.Body.Bytes(), &resp)
	expected := "Invalid format. Use ?workload=name:duration"
	if resp["result"] != expected {
		t.Errorf("Expected error message %q, got %q", expected, resp["result"])
	}
}

func TestHandler_Workload_InvalidDuration(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/workload?work=job1:notaduration", nil)
	res := httptest.NewRecorder()

	h := handler.Workload()
	h(res, req, nil)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	var resp map[string]string
	err := json.Unmarshal(res.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %s", err.Error())
	}
	if !contains(resp["result"], "Invalid duration") {
		t.Errorf("Expected error message containing 'Invalid duration', got %q", res.Body.String())
	}
}

func TestHandler_Workload_ZeroDuration(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/workload?work=job1:0s", nil)
	res := httptest.NewRecorder()

	h := handler.Workload()
	h(res, req, nil)

	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.Code)
	}

	var resp map[string]string
	json.Unmarshal(res.Body.Bytes(), &resp)
	expected := "Invalid duration: Duration must be expressed as [duration][suffix] where suffix=ns, us(or µs), ms, s, m, h and must be >0"
	if resp["result"] != expected {
		t.Errorf("Expected error message %q, got %q", expected, resp["result"])
	}
}

// Helper function to check substring presence
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(s) == len(substr) && s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || s[1:len(substr)+1] == substr))
}

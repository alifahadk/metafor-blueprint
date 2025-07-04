package handlers

import (
	"net/http"
	"rimcs/metafor-blueprint/workerpool"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

func (h Handler) Workload() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		workloadQueryParam := r.URL.Query().Get("work")
		if workloadQueryParam == "" {
			writeResponse(w, http.StatusBadRequest, "result", "Missing workload parameter")
			return
		}

		parts := strings.SplitN(workloadQueryParam, ":", 2)
		if len(parts) != 2 {
			writeResponse(w, http.StatusBadRequest, "result", "Invalid format. Use ?workload=name:duration")
			return
		}

		name := parts[0]
		duration, err := time.ParseDuration(parts[1])
		if err != nil {
			writeResponse(w, http.StatusBadRequest, "result", "Invalid duration "+err.Error())
			return
		}
		if duration.Nanoseconds() == 0 {
			writeResponse(w, http.StatusBadRequest, "result", "Invalid duration: Duration must be expressed as [duration][suffix] where suffix=ns, us(or µs), ms, s, m, h and must be >0")
			return
		}

		done := make(chan struct{})
		job := workerpool.Job{Name: name, Duration: duration, Done: done}

		select {
		case h.JobQueue <- job:
			<-done // Wait for job to complete
			writeResponse(w, http.StatusOK, "result", "success")
		default:
			writeResponse(w, http.StatusServiceUnavailable, "result", "unavailable")
		}
	}
}

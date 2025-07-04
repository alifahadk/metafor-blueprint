package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func writeResponse(w http.ResponseWriter, status int, key, value string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]string{key: value}
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to encode response: %v"}`, err), http.StatusInternalServerError)
		return err
	}

	_, err = w.Write(data)
	return err
}

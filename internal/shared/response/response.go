package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard API JSON response format.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSON writes a successful JSON response.
func JSON(w http.ResponseWriter, status int, message string, data interface{}) {
	write(w, status, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error writes an error JSON response.
func Error(w http.ResponseWriter, status int, message, code string) {
	write(w, status, Envelope{
		Success: false,
		Message: message,
		Error:   code,
	})
}

func write(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"success":false,"message":"failed to encode response","error":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
	}
}

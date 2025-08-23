package commons

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse is the standard JSON response for errors.
type ErrorResponse struct {
	Message string `json:"message,omitempty"`
	// Use interface{} to allow for different error structures (string or []FieldError).
	Error interface{} `json:"error"`
}

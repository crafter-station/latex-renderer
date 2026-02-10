package handler

// ErrorResponse represents an API error.
type ErrorResponse struct {
	Error  string `json:"error" example:"latex render failed"`
	Detail string `json:"detail,omitempty" example:"Undefined control sequence"`
}

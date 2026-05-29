package response

// Envelope is the standard response structure for all API endpoints
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Error   *ErrorInfo  `json:"error"`
	Meta    *Meta       `json:"meta"`
}

type ErrorInfo struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type Meta struct {
	Page     int   `json:"page,omitempty"`
	PageSize int   `json:"page_size,omitempty"`
	Total    int64 `json:"total,omitempty"`
}

// Success returns a successful response with data
func Success(data interface{}, message string) *Envelope {
	return &Envelope{
		Success: true,
		Data:    data,
		Message: message,
		Error:   nil,
	}
}

// SuccessWithMeta returns a successful response with metadata (for pagination)
func SuccessWithMeta(data interface{}, message string, meta *Meta) *Envelope {
	return &Envelope{
		Success: true,
		Data:    data,
		Message: message,
		Error:   nil,
		Meta:    meta,
	}
}

// Error returns an error response
func Error(code, message string, details []ErrorDetail) *Envelope {
	return &Envelope{
		Success: false,
		Data:    nil,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// ErrorSimple returns an error response without details
func ErrorSimple(code, message string) *Envelope {
	return Error(code, message, nil)
}

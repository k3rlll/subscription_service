package errDTO

// ErrorDTO represents the structure of an error response returned by the API
// It is written that way to provide a consistent format for error responses across the application
type ErrorDTO struct {
	HTTPStatus int    `json:"-"` //wont be included in JSON response
	Code       string `json:"error"`
	Message    string `json:"message"`
}

// Error implements the error interface, allowing ErrorDTO to be used as an error type in Go
func (e *ErrorDTO) Error() string {
	return e.Message
}

// New is a helper function to create a new ErrorDTO instance with the provided HTTP status, error code, and message
func New(status int, code string, message string) *ErrorDTO {
	return &ErrorDTO{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
	}
}

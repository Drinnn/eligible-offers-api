package errors

import "net/http"

type HttpError struct {
	StatusCode int               `json:"statusCode"`
	Message    string            `json:"message"`
	Errors     map[string]string `json:"errors,omitempty"`
}

func (e *HttpError) Error() string {
	return e.Message
}

func NewBadRequestError(message string, errors map[string]string) error {
	return &HttpError{StatusCode: http.StatusBadRequest, Message: message, Errors: errors}
}

func NewInternalServerError(message string) error {
	return &HttpError{StatusCode: http.StatusInternalServerError, Message: message}
}

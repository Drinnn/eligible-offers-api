package helpers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Errors map[string]string `json:"errors"`
}

func FormatValidationErrors(err error) ErrorResponse {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			fieldName := e.Field()
			errors[fieldName] = getErrorMessage(e)
		}
	}

	return ErrorResponse{Errors: errors}
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "gt":
		return e.Field() + " must be greater than " + e.Param()
	case "len":
		return e.Field() + " must be exactly " + e.Param() + " characters"
	case "numeric":
		return e.Field() + " must contain only numeric characters"
	default:
		return e.Field() + " is invalid"
	}
}

func WriteJSONError(w http.ResponseWriter, statusCode int, errorResponse ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

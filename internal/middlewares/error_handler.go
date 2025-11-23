package middlewares

import (
	"encoding/json"
	"log"
	"net/http"

	httpErrors "github.com/Drinnn/eligible-offers-api/internal/errors"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func ErrorHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			switch e := err.(type) {
			case *httpErrors.HttpError:
				w.WriteHeader(e.StatusCode)
				json.NewEncoder(w).Encode(e)
			case *httpErrors.ServiceError:
				log.Printf("Service error: %v", e)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"message": e.Error(),
				})
			default:
				log.Printf("Unknown error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"message": "Internal server error",
				})
			}
		}
	}
}

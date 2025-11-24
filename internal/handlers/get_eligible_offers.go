package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Drinnn/eligible-offers-api/internal/dtos"
	httpErrors "github.com/Drinnn/eligible-offers-api/internal/errors"
	"github.com/Drinnn/eligible-offers-api/internal/use_cases"
	"github.com/go-chi/chi"
)

type GetEligibleOffersHandler struct {
	getEligibleOffersUseCase *use_cases.GetEligibleOffersUseCase
}

func NewGetEligibleOffersHandler(getEligibleOffersUseCase *use_cases.GetEligibleOffersUseCase) *GetEligibleOffersHandler {
	return &GetEligibleOffersHandler{
		getEligibleOffersUseCase: getEligibleOffersUseCase,
	}
}

func (h *GetEligibleOffersHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	userID := chi.URLParam(r, "user_id")

	nowStr := r.URL.Query().Get("now")
	now := time.Now()
	if nowStr != "" {
		parsed, err := time.Parse(time.RFC3339, nowStr)
		if err != nil {
			return httpErrors.NewBadRequestError("Invalid now parameter", map[string]string{
				"now": "invalid time format. expected RFC3339 timestamp",
			})
		}
		now = parsed
	}

	result, err := h.getEligibleOffersUseCase.Execute(&dtos.GetEligibleOffersRequest{
		UserID: userID,
		Now:    now,
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	return nil
}

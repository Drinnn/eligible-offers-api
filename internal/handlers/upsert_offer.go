package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Drinnn/eligible-offers-api/internal/dtos"
	httpErrors "github.com/Drinnn/eligible-offers-api/internal/errors"
	"github.com/Drinnn/eligible-offers-api/internal/helpers"
	"github.com/Drinnn/eligible-offers-api/internal/use_cases"
)

type UpsertOfferHandler struct {
	upsertOfferUseCase *use_cases.UpsertOfferUseCase
}

func NewUpsertOfferHandler(upsertOfferUseCase *use_cases.UpsertOfferUseCase) *UpsertOfferHandler {
	return &UpsertOfferHandler{
		upsertOfferUseCase: upsertOfferUseCase,
	}
}

func (h *UpsertOfferHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	var request dtos.UpsertOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return httpErrors.NewBadRequestError("Invalid request body", nil)
	}

	if err := request.Validate(); err != nil {
		errorResponse := helpers.FormatValidationErrors(err)
		return httpErrors.NewBadRequestError("Invalid request body", errorResponse.Errors)
	}

	offer, err := h.upsertOfferUseCase.Execute(&request)
	if err != nil {
		return err
	}

	response := dtos.UpsertOfferResponse{
		ID:           offer.ID,
		MerchantID:   offer.MerchantID,
		MCCWhitelist: offer.MCCWhitelist,
		Active:       offer.Active,
		MinTxnCount:  offer.MinTxnCount,
		LookbackDays: offer.LookbackDays,
		StartsAt:     offer.StartsAt,
		EndsAt:       offer.EndsAt,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	return nil
}

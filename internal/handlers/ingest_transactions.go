package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Drinnn/eligible-offers-api/internal/dtos"
	httpErrors "github.com/Drinnn/eligible-offers-api/internal/errors"
	"github.com/Drinnn/eligible-offers-api/internal/helpers"
	"github.com/Drinnn/eligible-offers-api/internal/use_cases"
)

type IngestTransactionsHandler struct {
	ingestTransactionsUseCase *use_cases.IngestTransactionsUseCase
}

func NewIngestTransactionsHandler(ingestTransactionsUseCase *use_cases.IngestTransactionsUseCase) *IngestTransactionsHandler {
	return &IngestTransactionsHandler{
		ingestTransactionsUseCase: ingestTransactionsUseCase,
	}
}

func (h *IngestTransactionsHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	var request dtos.IngestTransactionsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return httpErrors.NewBadRequestError("Invalid request body", nil)
	}

	if err := request.Validate(); err != nil {
		errorResponse := helpers.FormatValidationErrors(err)
		return httpErrors.NewBadRequestError("Invalid request body", errorResponse.Errors)
	}

	response, err := h.ingestTransactionsUseCase.Execute(&request)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	return nil
}

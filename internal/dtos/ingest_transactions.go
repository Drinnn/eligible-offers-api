package dtos

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type IngestTransactionsRequest struct {
	Transactions []TransactionDto `json:"transactions" validate:"required,dive"`
}

type TransactionDto struct {
	ID          string    `json:"id" validate:"required"`
	UserID      string    `json:"user_id" validate:"required"`
	MerchantID  string    `json:"merchant_id" validate:"required"`
	MCC         string    `json:"mcc" validate:"required,len=4,numeric"`
	AmountCents int64     `json:"amount_cents" validate:"required,gt=0"`
	ApprovedAt  time.Time `json:"approved_at" validate:"required"`
}

func (r *IngestTransactionsRequest) Validate() error {
	validator := validator.New()
	return validator.Struct(r)
}

type IngestTransactionsResponse struct {
	Inserted int `json:"inserted"`
}

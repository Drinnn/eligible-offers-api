package dtos

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type UpsertOfferRequest struct {
	ID           string    `json:"id"`
	MerchantID   string    `json:"merchant_id" validate:"required"`
	MCCWhitelist []string  `json:"mcc_whitelist" validate:"required,dive,len=4,numeric"`
	Active       bool      `json:"active"`
	MinTxnCount  int       `json:"min_txn_count" validate:"required,gt=0"`
	LookbackDays int       `json:"lookback_days" validate:"required,gt=0"`
	StartsAt     time.Time `json:"starts_at" validate:"required"`
	EndsAt       time.Time `json:"ends_at" validate:"required"`
}

func (r *UpsertOfferRequest) Validate() error {
	validator := validator.New()
	return validator.Struct(r)
}

type UpsertOfferResponse struct {
	ID           string    `json:"id"`
	MerchantID   string    `json:"merchant_id"`
	MCCWhitelist []string  `json:"mcc_whitelist"`
	Active       bool      `json:"active"`
	MinTxnCount  int       `json:"min_txn_count"`
	LookbackDays int       `json:"lookback_days"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
}

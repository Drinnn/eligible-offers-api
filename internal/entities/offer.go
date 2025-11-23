package entities

import (
	"time"

	"github.com/google/uuid"
)

type Offer struct {
	ID           string   // uuid
	MerchantID   string   // uuid
	MCCWhitelist []string // e.g. ["5812", "5814"]
	Active       bool
	MinTxnCount  int       // N
	LookbackDays int       // K days
	StartsAt     time.Time // RFC3339 timestamp
	EndsAt       time.Time // RFC3339 timestamp
}

func NewOffer(id, merchantID string, mccWhitelist []string, active bool, minTxnCount, lookbackDays int, startsAt, endsAt time.Time) *Offer {
	if id == "" {
		id = uuid.New().String()
	}

	return &Offer{
		ID:           id,
		MerchantID:   merchantID,
		MCCWhitelist: mccWhitelist,
		Active:       active,
		MinTxnCount:  minTxnCount,
		LookbackDays: lookbackDays,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
	}
}

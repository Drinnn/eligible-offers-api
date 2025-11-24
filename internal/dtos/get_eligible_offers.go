package dtos

import "time"

type GetEligibleOffersRequest struct {
	UserID string
	Now    time.Time
}

type GetEligibleOffersResponse struct {
	UserID         string             `json:"user_id"`
	EligibleOffers []EligibleOfferDto `json:"eligible_offers"`
}

type EligibleOfferDto struct {
	OfferID string `json:"offer_id"`
	Reason  string `json:"reason"`
}

package use_cases

import (
	"github.com/Drinnn/eligible-offers-api/internal/dtos"
	"github.com/Drinnn/eligible-offers-api/internal/entities"
	customErrors "github.com/Drinnn/eligible-offers-api/internal/errors"
	"github.com/Drinnn/eligible-offers-api/internal/repositories"
)

type UpsertOfferUseCase struct {
	offerRepository repositories.OfferRepository
}

func NewUpsertOfferUseCase(offerRepository repositories.OfferRepository) *UpsertOfferUseCase {
	return &UpsertOfferUseCase{
		offerRepository: offerRepository,
	}
}

func (u *UpsertOfferUseCase) Execute(request *dtos.UpsertOfferRequest) (*entities.Offer, error) {
	if !request.StartsAt.Before(request.EndsAt) {
		return nil, customErrors.NewBadRequestError("starts_at must be before ends_at", nil)
	}

	offer := entities.NewOffer(request.ID, request.MerchantID, request.MCCWhitelist, request.Active, request.MinTxnCount, request.LookbackDays, request.StartsAt, request.EndsAt)

	if err := u.offerRepository.Upsert(offer); err != nil {
		return nil, customErrors.NewServiceError("failed to upsert offer")
	}

	return offer, nil
}

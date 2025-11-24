package use_cases

import (
	"fmt"
	"slices"
	"time"

	"github.com/Drinnn/eligible-offers-api/internal/dtos"
	"github.com/Drinnn/eligible-offers-api/internal/entities"
	customErrors "github.com/Drinnn/eligible-offers-api/internal/errors"
	"github.com/Drinnn/eligible-offers-api/internal/repositories"
)

type GetEligibleOffersUseCase struct {
	offerRepository       repositories.OfferRepository
	transactionRepository repositories.TransactionRepository
}

func NewGetEligibleOffersUseCase(offerRepository repositories.OfferRepository, transactionRepository repositories.TransactionRepository) *GetEligibleOffersUseCase {
	return &GetEligibleOffersUseCase{
		offerRepository:       offerRepository,
		transactionRepository: transactionRepository,
	}
}

func (u *GetEligibleOffersUseCase) Execute(request *dtos.GetEligibleOffersRequest) (*dtos.GetEligibleOffersResponse, error) {
	offers, err := u.offerRepository.GetAll()
	if err != nil {
		return nil, customErrors.NewServiceError("failed to get active offers")
	}

	userTransactions, err := u.transactionRepository.GetByUserID(request.UserID)
	if err != nil {
		return nil, customErrors.NewServiceError("failed to get user transactions")
	}

	activeOffers := u.filterActiveOffers(offers, request.Now)
	eligibleOffers := make([]*entities.Offer, 0)

	for _, offer := range activeOffers {
		lookbackStart := request.Now.AddDate(0, 0, -offer.LookbackDays)
		count := 0

		for _, transaction := range userTransactions {
			if transaction.ApprovedAt.Before(lookbackStart) || transaction.ApprovedAt.After(request.Now) {
				continue
			}

			if transaction.MerchantID == offer.MerchantID || slices.Contains(offer.MCCWhitelist, transaction.MCC) {
				count++
			}
		}

		if count >= offer.MinTxnCount {
			eligibleOffers = append(eligibleOffers, offer)
		}
	}

	eligibleOffersDtos := make([]dtos.EligibleOfferDto, 0)
	for _, offer := range eligibleOffers {
		eligibleOffersDtos = append(eligibleOffersDtos, dtos.EligibleOfferDto{
			OfferID: offer.ID,
			Reason:  fmt.Sprintf(">= %d transactions in last %d days", offer.MinTxnCount, offer.LookbackDays),
		})
	}

	return &dtos.GetEligibleOffersResponse{
		UserID:         request.UserID,
		EligibleOffers: eligibleOffersDtos,
	}, nil
}

func (u *GetEligibleOffersUseCase) filterActiveOffers(offers []*entities.Offer, now time.Time) []*entities.Offer {
	active := make([]*entities.Offer, 0)
	for _, offer := range offers {
		if offer.Active &&
			!now.Before(offer.StartsAt) &&
			!now.After(offer.EndsAt) {
			active = append(active, offer)
		}
	}
	return active
}

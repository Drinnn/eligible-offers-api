package repositories

import (
	"sync"

	"github.com/Drinnn/eligible-offers-api/internal/entities"
)

type OfferRepository interface {
	Upsert(offer *entities.Offer) error
	GetAll() ([]*entities.Offer, error)
}

type InMemoryOfferRepository struct {
	mu     sync.RWMutex
	offers map[string]*entities.Offer
}

func NewInMemoryOfferRepository() *InMemoryOfferRepository {
	return &InMemoryOfferRepository{
		offers: make(map[string]*entities.Offer),
	}
}

func (r *InMemoryOfferRepository) Upsert(offer *entities.Offer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.offers[offer.ID] = offer

	return nil
}

func (r *InMemoryOfferRepository) GetAll() ([]*entities.Offer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allOffers := make([]*entities.Offer, 0)
	for _, offer := range r.offers {
		allOffers = append(allOffers, offer)
	}
	return allOffers, nil
}

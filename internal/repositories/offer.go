package repositories

import (
	"sync"

	"github.com/Drinnn/eligible-offers-api/internal/entities"
)

type OfferRepository interface {
	Upsert(offer *entities.Offer) error
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

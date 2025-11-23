package repositories

import (
	"sync"

	"github.com/Drinnn/eligible-offers-api/internal/entities"
)

type TransactionRepository interface {
	Insert(transactions []*entities.Transaction) (int, error)
}

type InMemoryTransactionRepository struct {
	mu           sync.RWMutex
	transactions map[string]*entities.Transaction
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[string]*entities.Transaction),
	}
}

func (r *InMemoryTransactionRepository) Insert(transactions []*entities.Transaction) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inserted := 0
	for _, transaction := range transactions {
		if _, exists := r.transactions[transaction.ID]; !exists {
			r.transactions[transaction.ID] = transaction
			inserted++
		}
	}

	return inserted, nil
}

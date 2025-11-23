package use_cases

import (
	"github.com/Drinnn/eligible-offers-api/internal/dtos"
	"github.com/Drinnn/eligible-offers-api/internal/entities"
	customErrors "github.com/Drinnn/eligible-offers-api/internal/errors"
	"github.com/Drinnn/eligible-offers-api/internal/repositories"
)

type IngestTransactionsUseCase struct {
	transactionRepository repositories.TransactionRepository
}

func NewIngestTransactionsUseCase(transactionRepository repositories.TransactionRepository) *IngestTransactionsUseCase {
	return &IngestTransactionsUseCase{
		transactionRepository: transactionRepository,
	}
}

func (u *IngestTransactionsUseCase) Execute(request *dtos.IngestTransactionsRequest) (*dtos.IngestTransactionsResponse, error) {
	transactions := make([]*entities.Transaction, len(request.Transactions))
	for i, transaction := range request.Transactions {
		transactions[i] = entities.NewTransaction(transaction.ID, transaction.UserID, transaction.MerchantID, transaction.MCC, transaction.AmountCents, transaction.ApprovedAt)
	}

	inserted, err := u.transactionRepository.Insert(transactions)
	if err != nil {
		return nil, customErrors.NewServiceError("failed to ingest transactions")
	}

	return &dtos.IngestTransactionsResponse{
		Inserted: inserted,
	}, nil
}

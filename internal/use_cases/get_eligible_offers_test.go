package use_cases_test

import (
	"testing"
	"time"

	"github.com/Drinnn/eligible-offers-api/internal/dtos"
	"github.com/Drinnn/eligible-offers-api/internal/entities"
	"github.com/Drinnn/eligible-offers-api/internal/repositories"
	"github.com/Drinnn/eligible-offers-api/internal/use_cases"
)

func TestGetEligibleOffers_UserQualifies(t *testing.T) {
	offerRepo := repositories.NewInMemoryOfferRepository()
	txnRepo := repositories.NewInMemoryTransactionRepository()
	useCase := use_cases.NewGetEligibleOffersUseCase(offerRepo, txnRepo)

	// Given: An active offer with min txn count of 3 in last 30 days
	now := time.Date(2025, 10, 21, 10, 0, 0, 0, time.UTC)
	offer := &entities.Offer{
		ID:           "offer-1",
		MerchantID:   "merchant-1",
		Active:       true,
		MinTxnCount:  3,
		LookbackDays: 30,
		StartsAt:     now.AddDate(0, 0, -10), // 10 days ago
		EndsAt:       now.AddDate(0, 0, 10),  // 10 days in the future
	}
	offerRepo.Upsert(offer)

	// And: A user with 3 transactions in the last 30 days
	userID := "user-1"
	transactions := []*entities.Transaction{
		{ID: "txn-1", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -5)},
		{ID: "txn-2", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -10)},
		{ID: "txn-3", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -15)},
	}
	txnRepo.Insert(transactions)

	// When: We call the use case
	request := &dtos.GetEligibleOffersRequest{
		UserID: userID,
		Now:    now,
	}
	result, err := useCase.Execute(request)

	// Then: User must be eligible for the offer
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.EligibleOffers) != 1 {
		t.Fatalf("expected 1 eligible offer, got %d", len(result.EligibleOffers))
	}
	if result.EligibleOffers[0].OfferID != "offer-1" {
		t.Errorf("expected offer-1, got %s", result.EligibleOffers[0].OfferID)
	}
	expectedReason := ">= 3 transactions in last 30 days"
	if result.EligibleOffers[0].Reason != expectedReason {
		t.Errorf("expected reason '%s', got '%s'", expectedReason, result.EligibleOffers[0].Reason)
	}
}

func TestGetEligibleOffers_NotEnoughTransactions(t *testing.T) {
	offerRepo := repositories.NewInMemoryOfferRepository()
	txnRepo := repositories.NewInMemoryTransactionRepository()
	useCase := use_cases.NewGetEligibleOffersUseCase(offerRepo, txnRepo)

	// Given: An active offer requiring 3 transactions
	now := time.Date(2025, 10, 21, 10, 0, 0, 0, time.UTC)
	offer := &entities.Offer{
		ID:           "offer-1",
		MerchantID:   "merchant-1",
		Active:       true,
		MinTxnCount:  3,
		LookbackDays: 30,
		StartsAt:     now.AddDate(0, 0, -10),
		EndsAt:       now.AddDate(0, 0, 10),
	}
	offerRepo.Upsert(offer)

	// And: A user with only 2 transactions (not enough!)
	userID := "user-1"
	transactions := []*entities.Transaction{
		{ID: "txn-1", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -5)},
		{ID: "txn-2", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -10)},
	}
	txnRepo.Insert(transactions)

	// When: We call the use case
	request := &dtos.GetEligibleOffersRequest{
		UserID: userID,
		Now:    now,
	}
	result, err := useCase.Execute(request)

	// Then: User must NOT be eligible
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.EligibleOffers) != 0 {
		t.Errorf("expected 0 eligible offers, got %d", len(result.EligibleOffers))
	}
}

func TestGetEligibleOffers_OfferInactive(t *testing.T) {
	offerRepo := repositories.NewInMemoryOfferRepository()
	txnRepo := repositories.NewInMemoryTransactionRepository()
	useCase := use_cases.NewGetEligibleOffersUseCase(offerRepo, txnRepo)

	// Given: An INACTIVE offer
	now := time.Date(2025, 10, 21, 10, 0, 0, 0, time.UTC)
	offer := &entities.Offer{
		ID:           "offer-1",
		MerchantID:   "merchant-1",
		Active:       false, // Inactive!
		MinTxnCount:  3,
		LookbackDays: 30,
		StartsAt:     now.AddDate(0, 0, -10),
		EndsAt:       now.AddDate(0, 0, 10),
	}
	offerRepo.Upsert(offer)

	// And: A user with enough transactions
	userID := "user-1"
	transactions := []*entities.Transaction{
		{ID: "txn-1", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -5)},
		{ID: "txn-2", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -10)},
		{ID: "txn-3", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -15)},
	}
	txnRepo.Insert(transactions)

	// When: We call the use case
	request := &dtos.GetEligibleOffersRequest{
		UserID: userID,
		Now:    now,
	}
	result, err := useCase.Execute(request)

	// Then: User must NOT be eligible (offer is inactive)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.EligibleOffers) != 0 {
		t.Errorf("expected 0 eligible offers (inactive), got %d", len(result.EligibleOffers))
	}
}

func TestGetEligibleOffers_OutsideDateRange(t *testing.T) {
	offerRepo := repositories.NewInMemoryOfferRepository()
	txnRepo := repositories.NewInMemoryTransactionRepository()
	useCase := use_cases.NewGetEligibleOffersUseCase(offerRepo, txnRepo)

	// Given: An offer that has already EXPIRED
	now := time.Date(2025, 10, 21, 10, 0, 0, 0, time.UTC)
	offer := &entities.Offer{
		ID:           "offer-1",
		MerchantID:   "merchant-1",
		Active:       true,
		MinTxnCount:  3,
		LookbackDays: 30,
		StartsAt:     now.AddDate(0, 0, -60), // 60 days ago
		EndsAt:       now.AddDate(0, 0, -30), // Expired 30 days ago!
	}
	offerRepo.Upsert(offer)

	// And: A user with enough transactions
	userID := "user-1"
	transactions := []*entities.Transaction{
		{ID: "txn-1", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -5)},
		{ID: "txn-2", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -10)},
		{ID: "txn-3", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -15)},
	}
	txnRepo.Insert(transactions)

	// When: We call the use case
	request := &dtos.GetEligibleOffersRequest{
		UserID: userID,
		Now:    now,
	}
	result, err := useCase.Execute(request)

	// Then: User must NOT be eligible (offer expired)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.EligibleOffers) != 0 {
		t.Errorf("expected 0 eligible offers (expired), got %d", len(result.EligibleOffers))
	}
}

func TestGetEligibleOffers_TransactionsOutsideLookbackWindow(t *testing.T) {
	offerRepo := repositories.NewInMemoryOfferRepository()
	txnRepo := repositories.NewInMemoryTransactionRepository()
	useCase := use_cases.NewGetEligibleOffersUseCase(offerRepo, txnRepo)

	// Given: An active offer with 30 days lookback
	now := time.Date(2025, 10, 21, 10, 0, 0, 0, time.UTC)
	offer := &entities.Offer{
		ID:           "offer-1",
		MerchantID:   "merchant-1",
		Active:       true,
		MinTxnCount:  3,
		LookbackDays: 30, // Last 30 days only
		StartsAt:     now.AddDate(0, 0, -60),
		EndsAt:       now.AddDate(0, 0, 10),
	}
	offerRepo.Upsert(offer)

	// And: A user with transactions that are TOO OLD (outside lookback window)
	userID := "user-1"
	transactions := []*entities.Transaction{
		{ID: "txn-1", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -40)}, // 40 days ago
		{ID: "txn-2", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -50)}, // 50 days ago
		{ID: "txn-3", UserID: userID, MerchantID: "merchant-1", ApprovedAt: now.AddDate(0,
			0, -60)}, // 60 days ago
	}
	txnRepo.Insert(transactions)

	// When: We call the use case
	request := &dtos.GetEligibleOffersRequest{
		UserID: userID,
		Now:    now,
	}
	result, err := useCase.Execute(request)

	// Then: User must NOT be eligible (transactions too old)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.EligibleOffers) != 0 {
		t.Errorf("expected 0 eligible offers (txns too old), got %d", len(result.EligibleOffers))
	}
}

func TestGetEligibleOffers_MatchByMCC(t *testing.T) {
	offerRepo := repositories.NewInMemoryOfferRepository()
	txnRepo := repositories.NewInMemoryTransactionRepository()
	useCase := use_cases.NewGetEligibleOffersUseCase(offerRepo, txnRepo)

	// Given: An active offer matching by MCC whitelist
	now := time.Date(2025, 10, 21, 10, 0, 0, 0, time.UTC)
	offer := &entities.Offer{
		ID:           "offer-1",
		MerchantID:   "merchant-1",
		MCCWhitelist: []string{"5812", "5814"}, // Restaurant MCCs
		Active:       true,
		MinTxnCount:  2,
		LookbackDays: 30,
		StartsAt:     now.AddDate(0, 0, -10),
		EndsAt:       now.AddDate(0, 0, 10),
	}
	offerRepo.Upsert(offer)

	// And: A user with transactions at DIFFERENT merchants but matching MCCs
	userID := "user-1"
	transactions := []*entities.Transaction{
		{ID: "txn-1", UserID: userID, MerchantID: "merchant-2", MCC: "5812", ApprovedAt: now.AddDate(0,
			0, -5)}, // Different merchant, but matching MCC
		{ID: "txn-2", UserID: userID, MerchantID: "merchant-3", MCC: "5814", ApprovedAt: now.AddDate(0,
			0, -10)}, // Different merchant, but matching MCC
	}
	txnRepo.Insert(transactions)

	// When: We call the use case
	request := &dtos.GetEligibleOffersRequest{
		UserID: userID,
		Now:    now,
	}
	result, err := useCase.Execute(request)

	// Then: User must be eligible (matched by MCC)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.EligibleOffers) != 1 {
		t.Fatalf("expected 1 eligible offer, got %d", len(result.EligibleOffers))
	}
	if result.EligibleOffers[0].OfferID != "offer-1" {
		t.Errorf("expected offer-1, got %s", result.EligibleOffers[0].OfferID)
	}
	expectedReason := ">= 2 transactions in last 30 days"
	if result.EligibleOffers[0].Reason != expectedReason {
		t.Errorf("expected reason '%s', got '%s'", expectedReason, result.EligibleOffers[0].Reason)
	}
}

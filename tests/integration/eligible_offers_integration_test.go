package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Drinnn/eligible-offers-api/internal/handlers"
	"github.com/Drinnn/eligible-offers-api/internal/middlewares"
	"github.com/Drinnn/eligible-offers-api/internal/repositories"
	"github.com/Drinnn/eligible-offers-api/internal/use_cases"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

// setupTestServer creates a test HTTP server with all routes configured
func setupTestServer() *httptest.Server {
	// Initialize repositories (shared between all use cases)
	offerRepository := repositories.NewInMemoryOfferRepository()
	transactionRepository := repositories.NewInMemoryTransactionRepository()

	// Initialize use cases
	upsertOfferUseCase := use_cases.NewUpsertOfferUseCase(offerRepository)
	ingestTransactionsUseCase := use_cases.NewIngestTransactionsUseCase(transactionRepository)
	getEligibleOffersUseCase := use_cases.NewGetEligibleOffersUseCase(offerRepository, transactionRepository)

	// Initialize handlers
	upsertOfferHandler := handlers.NewUpsertOfferHandler(upsertOfferUseCase)
	ingestTransactionsHandler := handlers.NewIngestTransactionsHandler(ingestTransactionsUseCase)
	getEligibleOffersHandler := handlers.NewGetEligibleOffersHandler(getEligibleOffersUseCase)

	// Setup router
	router := chi.NewRouter()
	router.Use(middlewares.JSON)
	router.Use(middleware.Logger)

	router.Post("/offers", middlewares.ErrorHandler(upsertOfferHandler.Handle))
	router.Post("/transactions", middlewares.ErrorHandler(ingestTransactionsHandler.Handle))
	router.Get("/users/{user_id}/eligible-offers", middlewares.ErrorHandler(getEligibleOffersHandler.Handle))

	// Create test server
	return httptest.NewServer(router)
}

func TestEligibleOffersIntegration_UserQualifies(t *testing.T) {
	// Given: A test server
	server := setupTestServer()
	defer server.Close()

	// When: We create an active offer
	offerPayload := `{
		"merchant_id": "merchant-123",
		"mcc_whitelist": ["5812", "5814"],
		"active": true,
		"min_txn_count": 3,
		"lookback_days": 30,
		"starts_at": "2025-01-01T00:00:00Z",
		"ends_at": "2025-12-31T23:59:59Z"
	}`

	resp, err := http.Post(server.URL+"/offers", "application/json", bytes.NewBuffer([]byte(offerPayload)))
	if err != nil {
		t.Fatalf("Failed to create offer: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	// Extract offer ID from response
	var offerResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&offerResp); err != nil {
		t.Fatalf("Failed to decode offer response: %v", err)
	}
	offerID := offerResp["id"].(string)

	// And: We ingest 3 transactions for a user matching the offer
	txnPayload := `{
		"transactions": [
			{
				"id": "txn-1",
				"user_id": "user-456",
				"merchant_id": "merchant-123",
				"mcc": "5812",
				"amount_cents": 1000,
				"approved_at": "2025-11-20T12:00:00Z"
			},
			{
				"id": "txn-2",
				"user_id": "user-456",
				"merchant_id": "merchant-123",
				"mcc": "5812",
				"amount_cents": 2000,
				"approved_at": "2025-11-21T12:00:00Z"
			},
			{
				"id": "txn-3",
				"user_id": "user-456",
				"merchant_id": "merchant-123",
				"mcc": "5812",
				"amount_cents": 1500,
				"approved_at": "2025-11-22T12:00:00Z"
			}
		]
	}`

	resp, err = http.Post(server.URL+"/transactions", "application/json", bytes.NewBuffer([]byte(txnPayload)))
	if err != nil {
		t.Fatalf("Failed to ingest transactions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	// When: We check eligibility for the user
	resp, err = http.Get(server.URL + "/users/user-456/eligible-offers?now=2025-11-23T10:00:00Z")
	if err != nil {
		t.Fatalf("Failed to get eligible offers: %v", err)
	}
	defer resp.Body.Close()

	// Then: User should be eligible for the offer
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var eligibleResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&eligibleResp); err != nil {
		t.Fatalf("Failed to decode eligible offers response: %v", err)
	}

	// Verify response structure
	if eligibleResp["user_id"] != "user-456" {
		t.Errorf("Expected user_id 'user-456', got '%v'", eligibleResp["user_id"])
	}

	eligibleOffers := eligibleResp["eligible_offers"].([]any)
	if len(eligibleOffers) != 1 {
		t.Fatalf("Expected 1 eligible offer, got %d", len(eligibleOffers))
	}

	// Verify the offer details
	offer := eligibleOffers[0].(map[string]any)
	if offer["offer_id"] != offerID {
		t.Errorf("Expected offer_id '%s', got '%v'", offerID, offer["offer_id"])
	}

	expectedReason := ">= 3 transactions in last 30 days"
	if offer["reason"] != expectedReason {
		t.Errorf("Expected reason '%s', got '%v'", expectedReason, offer["reason"])
	}
}

func TestEligibleOffersIntegration_UserDoesNotQualify(t *testing.T) {
	// Given: A test server
	server := setupTestServer()
	defer server.Close()

	// When: We create an active offer requiring 3 transactions
	offerPayload := `{
		"merchant_id": "merchant-789",
		"mcc_whitelist": ["5411"],
		"active": true,
		"min_txn_count": 3,
		"lookback_days": 30,
		"starts_at": "2025-01-01T00:00:00Z",
		"ends_at": "2025-12-31T23:59:59Z"
	}`

	resp, err := http.Post(server.URL+"/offers", "application/json", bytes.NewBuffer([]byte(offerPayload)))
	if err != nil {
		t.Fatalf("Failed to create offer: %v", err)
	}
	defer resp.Body.Close()

	// And: We ingest only 2 transactions (not enough!)
	txnPayload := `{
		"transactions": [
			{
				"id": "txn-10",
				"user_id": "user-999",
				"merchant_id": "merchant-789",
				"mcc": "5411",
				"amount_cents": 1000,
				"approved_at": "2025-11-20T12:00:00Z"
			},
			{
				"id": "txn-11",
				"user_id": "user-999",
				"merchant_id": "merchant-789",
				"mcc": "5411",
				"amount_cents": 2000,
				"approved_at": "2025-11-21T12:00:00Z"
			}
		]
	}`

	resp, err = http.Post(server.URL+"/transactions", "application/json", bytes.NewBuffer([]byte(txnPayload)))
	if err != nil {
		t.Fatalf("Failed to ingest transactions: %v", err)
	}
	defer resp.Body.Close()

	// When: We check eligibility
	resp, err = http.Get(server.URL + "/users/user-999/eligible-offers?now=2025-11-23T10:00:00Z")
	if err != nil {
		t.Fatalf("Failed to get eligible offers: %v", err)
	}
	defer resp.Body.Close()

	// Then: User should NOT be eligible
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var eligibleResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&eligibleResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	eligibleOffers := eligibleResp["eligible_offers"].([]any)
	if len(eligibleOffers) != 0 {
		t.Errorf("Expected 0 eligible offers, got %d", len(eligibleOffers))
	}
}

func TestEligibleOffersIntegration_MatchByMCC(t *testing.T) {
	// Given: A test server
	server := setupTestServer()
	defer server.Close()

	// When: We create an offer that matches by MCC
	offerPayload := `{
		"merchant_id": "merchant-original",
		"mcc_whitelist": ["5812", "5814"],
		"active": true,
		"min_txn_count": 2,
		"lookback_days": 30,
		"starts_at": "2025-01-01T00:00:00Z",
		"ends_at": "2025-12-31T23:59:59Z"
	}`

	resp, err := http.Post(server.URL+"/offers", "application/json", bytes.NewBuffer([]byte(offerPayload)))
	if err != nil {
		t.Fatalf("Failed to create offer: %v", err)
	}
	defer resp.Body.Close()

	// And: We ingest transactions at DIFFERENT merchants but matching MCCs
	txnPayload := `{
		"transactions": [
			{
				"id": "txn-20",
				"user_id": "user-777",
				"merchant_id": "merchant-different-1",
				"mcc": "5812",
				"amount_cents": 1000,
				"approved_at": "2025-11-20T12:00:00Z"
			},
			{
				"id": "txn-21",
				"user_id": "user-777",
				"merchant_id": "merchant-different-2",
				"mcc": "5814",
				"amount_cents": 2000,
				"approved_at": "2025-11-21T12:00:00Z"
			}
		]
	}`

	resp, err = http.Post(server.URL+"/transactions", "application/json", bytes.NewBuffer([]byte(txnPayload)))
	if err != nil {
		t.Fatalf("Failed to ingest transactions: %v", err)
	}
	defer resp.Body.Close()

	// When: We check eligibility
	resp, err = http.Get(server.URL + "/users/user-777/eligible-offers?now=2025-11-23T10:00:00Z")
	if err != nil {
		t.Fatalf("Failed to get eligible offers: %v", err)
	}
	defer resp.Body.Close()

	// Then: User should be eligible (matched by MCC, not merchant_id)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var eligibleResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&eligibleResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	eligibleOffers := eligibleResp["eligible_offers"].([]any)
	if len(eligibleOffers) != 1 {
		t.Fatalf("Expected 1 eligible offer (matched by MCC), got %d", len(eligibleOffers))
	}
}

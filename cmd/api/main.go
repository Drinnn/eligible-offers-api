package main

import (
	"log"
	"net/http"

	"github.com/Drinnn/eligible-offers-api/internal/handlers"
	"github.com/Drinnn/eligible-offers-api/internal/middlewares"
	"github.com/Drinnn/eligible-offers-api/internal/repositories"
	"github.com/Drinnn/eligible-offers-api/internal/use_cases"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	offerRepository := repositories.NewInMemoryOfferRepository()
	upsertOfferUseCase := use_cases.NewUpsertOfferUseCase(offerRepository)
	upsertOfferHandler := handlers.NewUpsertOfferHandler(upsertOfferUseCase)

	transactionRepository := repositories.NewInMemoryTransactionRepository()
	ingestTransactionsUseCase := use_cases.NewIngestTransactionsUseCase(transactionRepository)
	ingestTransactionsHandler := handlers.NewIngestTransactionsHandler(ingestTransactionsUseCase)

	getEligibleOffersUseCase := use_cases.NewGetEligibleOffersUseCase(offerRepository, transactionRepository)
	getEligibleOffersHandler := handlers.NewGetEligibleOffersHandler(getEligibleOffersUseCase)

	router := chi.NewRouter()
	router.Use(middlewares.JSON)
	router.Use(middleware.Logger)

	router.Post("/offers", middlewares.ErrorHandler(upsertOfferHandler.Handle))
	router.Post("/transactions", middlewares.ErrorHandler(ingestTransactionsHandler.Handle))
	router.Get("/users/{user_id}/eligible-offers", middlewares.ErrorHandler(getEligibleOffersHandler.Handle))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

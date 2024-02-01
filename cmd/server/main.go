package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/vancho-go/GoCurrencyGate/internal/app/currency"
	"github.com/vancho-go/GoCurrencyGate/internal/app/handlers"
	"github.com/vancho-go/GoCurrencyGate/internal/app/repository"
	"github.com/vancho-go/GoCurrencyGate/internal/config"
	"github.com/vancho-go/GoCurrencyGate/internal/pkg/logger"
)

var timeout = 5 * time.Second

func main() {
	logger, err := logger.NewLogger("info")
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	configs, err := config.BuildServer()
	if err != nil {
		logger.Fatal("error building server  configuration: %v", zap.Error(err))
	}

	dbInstance, err := repository.Initialize(configs.DatabaseURI)
	if err != nil {
		logger.Fatal("error initialising database", zap.Error(err))
	}

	// Создаем новый планировщик
	cbrFetcher := &handlers.CBRFetcher{}
	c := cron.New(cron.WithLocation(time.FixedZone("MSK", 3*60*60))) // Московское время (UTC+3)
	_, err = c.AddFunc("0 10 * * *", func() { currency.UpdateCurrencyRates(dbInstance, cbrFetcher, logger, timeout) })
	if err != nil {
		logger.Fatal("error adding cron job: %v", zap.Error(err))
	}
	c.Start()

	r := chi.NewRouter()

	r.Get("/currency", handlers.GetCurrencyRateHandler(dbInstance, cbrFetcher, logger))

	logger.Info("starting server", zap.String("address", configs.ServerRunAddress))
	err = http.ListenAndServe(configs.ServerRunAddress, r)
	if err != nil {
		logger.Fatal("error starting server: %v", zap.Error(err))
	}

}

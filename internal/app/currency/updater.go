package currency

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/vancho-go/GoCurrencyGate/internal/app/models"
	"github.com/vancho-go/GoCurrencyGate/internal/pkg/logger"
)

const dateFormat = "02/01/2006"

type RateSaver interface {
	SaveCurrencyRate(context.Context, *models.ValuteCurs, time.Time) error
}

type Fetcher interface {
	GetCurrency(string) (*models.ValuteCurs, error)
}

func UpdateCurrencyRates(cr RateSaver, cf Fetcher, logger logger.Logger, timeout time.Duration) {
	date := time.Now().Format(dateFormat)
	cursCBR, err := cf.GetCurrency(date)
	if err != nil {
		logger.Error("updateCurrencyRates: error updating currency rates", zap.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = cr.SaveCurrencyRate(ctx, cursCBR, time.Now())
	if err != nil {
		logger.Error("updateCurrencyRates: failed to save currency rates: ", zap.Error(err))
		return
	}
	logger.Info("updateCurrencyRates: currency rates updated successfully")
}

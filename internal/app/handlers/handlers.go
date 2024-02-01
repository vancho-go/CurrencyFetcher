package handlers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/html/charset"

	"github.com/vancho-go/GoCurrencyGate/internal/app/currency"
	"github.com/vancho-go/GoCurrencyGate/internal/app/models"
	"github.com/vancho-go/GoCurrencyGate/internal/app/repository"
	"github.com/vancho-go/GoCurrencyGate/internal/pkg/logger"
)

const dateFormat = "02/01/2006"

var cbrAPI = "https://www.cbr.ru/scripts/XML_daily.asp"

type CurrencyExchanger interface {
	GetCurrencyRate(context.Context, string, time.Time) (models.APIGetValuteResponse, error)
}

type CBRFetcher struct{}

func GetCurrencyRateHandler(ce CurrencyExchanger, cf currency.Fetcher, logger logger.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		date := req.URL.Query().Get("date")
		val := req.URL.Query().Get("val")

		if date == "" {
			date = time.Now().Format(dateFormat)
		}

		if val == "" {
			http.Error(res, "Invalid val format", http.StatusBadRequest)
			return
		}

		dateTime, err := time.Parse(dateFormat, date)
		if err != nil {
			http.Error(res, "Invalid data format", http.StatusBadRequest)
			return
		}

		cursLocal, err := ce.GetCurrencyRate(req.Context(), val, dateTime)
		if err == nil {
			// Успешное получение из БД
			if err = json.NewEncoder(res).Encode(cursLocal); err != nil {
				logger.Error("getCurrencyRateHandler:", zap.Error(err))
				http.Error(res, "Failed to encode response", http.StatusInternalServerError)
			}
			return
		} else if !errors.Is(err, repository.ErrNoValuteDataFound) {
			logger.Error("getCurrencyRateHandler:", zap.Error(err))
			http.Error(res, "Failed to retrieve currency rate", http.StatusInternalServerError)
			return
		}

		cursCBR, err := cf.GetCurrency(date)
		if err != nil {
			logger.Error("getCurrencyRateHandler:", zap.Error(err))
			http.Error(res, "Failed to retrieve currency rate from CBR", http.StatusInternalServerError)
			return
		}

		for _, v := range cursCBR.Valutes {
			if strings.ToLower(v.CharCode) == strings.ToLower(val) {
				valuteResponse := &models.APIGetValuteResponse{
					Name:     v.Name,
					CharCode: v.CharCode,
					Date:     date,
					Value:    strings.Replace(v.Value, ",", ".", -1),
					Nominal:  v.Nominal,
				}

				if err = json.NewEncoder(res).Encode(valuteResponse); err != nil {
					logger.Error("getCurrencyRateHandler:", zap.Error(err))
					http.Error(res, "Internal error", http.StatusInternalServerError)
				}
				return
			}
		}

		http.Error(res, "Valute not found", http.StatusNotFound)

	}
}

func (f *CBRFetcher) GetCurrency(date string) (*models.ValuteCurs, error) {
	url := fmt.Sprintf("%s?date_req=%s", cbrAPI, date)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("GetCurrency: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetCurrency: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetCurrency: CBR returned non-OK status: %d: %w", resp.StatusCode, err)
	}

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel

	valCurs := &models.ValuteCurs{}
	if err = decoder.Decode(valCurs); err != nil {
		return nil, fmt.Errorf("GetCurrency: %w", err)
	}

	return valCurs, nil
}

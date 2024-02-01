package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vancho-go/GoCurrencyGate/internal/app/models"
)

type MockCurrencyExchanger struct {
	mock.Mock
}

func (m *MockCurrencyExchanger) GetCurrencyRate(ctx context.Context, val string, date time.Time) (models.APIGetValuteResponse, error) {
	args := m.Called(ctx, val, date)
	return args.Get(0).(models.APIGetValuteResponse), args.Error(1)
}

type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) GetCurrency(date string) (*models.ValuteCurs, error) {
	args := m.Called(date)
	return args.Get(0).(*models.ValuteCurs), args.Error(1)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func TestGetCurrencyRateHandler(t *testing.T) {
	mockExchanger := new(MockCurrencyExchanger)
	mockFetcher := new(MockFetcher)
	mockLogger := new(MockLogger)

	handler := GetCurrencyRateHandler(mockExchanger, mockFetcher, mockLogger)

	t.Run("success", func(t *testing.T) {
		date := time.Now().Format(dateFormat)
		val := "USD"
		mockExchanger.On("GetCurrencyRate", mock.Anything, val, mock.AnythingOfType("time.Time")).
			Return(models.APIGetValuteResponse{Value: "1.2345"}, nil)

		req, err := http.NewRequest("GET", "/?date="+date+"&val="+val, nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockExchanger.AssertExpectations(t)
	})

	t.Run("invalid_date_format", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/?date=invalid&val=USD", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing_val", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/?date="+time.Now().Format(dateFormat), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestGetCurrency_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xmlResponse := `<ValuteCurs><Valute ID="R01235"><NumCode>840</NumCode><CharCode>USD</CharCode><Nominal>1</Nominal><Name>Доллар США</Name><Value>75,0000</Value></Valute></ValuteCurs>`
		io.WriteString(w, xmlResponse)
	}))
	defer ts.Close()

	cbrAPI = ts.URL
	fetcher := CBRFetcher{}

	valCurs, err := fetcher.GetCurrency("02/03/2021")

	require.NoError(t, err)
	assert.NotNil(t, valCurs)
	assert.Equal(t, "Доллар США", valCurs.Valutes[0].Name)
	assert.Equal(t, 1, valCurs.Valutes[0].Nominal)
	assert.Equal(t, "USD", valCurs.Valutes[0].CharCode)
	assert.Equal(t, "75,0000", valCurs.Valutes[0].Value)
}

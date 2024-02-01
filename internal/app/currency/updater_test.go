package currency

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/vancho-go/GoCurrencyGate/internal/app/models"

	"github.com/stretchr/testify/mock"
)

type MockRateSaver struct {
	mock.Mock
}

func (m *MockRateSaver) SaveCurrencyRate(ctx context.Context, valuteCurs *models.ValuteCurs, date time.Time) error {
	args := m.Called(ctx, valuteCurs, date)
	return args.Error(0)
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

func TestUpdateCurrencyRates(t *testing.T) {
	mockRateSaver := new(MockRateSaver)
	mockFetcher := new(MockFetcher)
	mockLogger := new(MockLogger)

	cursCBR := &models.ValuteCurs{}
	date := time.Now().Format(dateFormat)
	timeout := 5 * time.Second

	mockFetcher.On("GetCurrency", date).Return(cursCBR, nil)
	mockRateSaver.On("SaveCurrencyRate", mock.AnythingOfType("*context.timerCtx"), cursCBR, mock.AnythingOfType("time.Time")).Return(nil)
	mockLogger.On("Info", mock.AnythingOfType("string"), mock.Anything)

	UpdateCurrencyRates(mockRateSaver, mockFetcher, mockLogger, timeout)

	mockFetcher.AssertExpectations(t)
	mockRateSaver.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

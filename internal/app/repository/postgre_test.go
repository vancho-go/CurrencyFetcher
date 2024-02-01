package repository

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/vancho-go/GoCurrencyGate/internal/app/models"
)

func TestSaveCurrencyRate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	myDB := &DB{db}
	ctx := context.Background()
	date := time.Now()

	valuteCurs := &models.ValuteCurs{
		Valutes: []models.Valute{
			{CharCode: "USD", Nominal: 1, Name: "Доллар США", Value: "75,0000"},
			{CharCode: "EUR", Nominal: 1, Name: "Евро", Value: "90,0000"},
		},
	}

	mock.ExpectBegin()
	stmt := mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO currency_rates (char_code, nominal, name, value, date) VALUES ($1, $2, $3, $4, $5)"))
	for _, valute := range valuteCurs.Valutes {
		valute.Value = strings.Replace(valute.Value, ",", ".", -1)
		stmt.ExpectExec().WithArgs(valute.CharCode, valute.Nominal, valute.Name, valute.Value, date).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err = myDB.SaveCurrencyRate(ctx, valuteCurs, date)
	require.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestSaveCurrencyRate_ErrorOnBegin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	myDB := &DB{db}
	ctx := context.Background()
	date := time.Now()
	valuteCurs := &models.ValuteCurs{}

	mock.ExpectBegin().WillReturnError(sql.ErrTxDone)

	err = myDB.SaveCurrencyRate(ctx, valuteCurs, date)
	require.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetCurrencyRate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	myDB := &DB{db}
	ctx := context.Background()
	charCode := "USD"
	date := time.Now()

	expectedValute := models.APIGetValuteResponse{
		CharCode: charCode,
		Name:     "Доллар США",
		Value:    "75.0000",
		Nominal:  1,
		Date:     date.Format("02/01/2006"),
	}

	rows := sqlmock.NewRows([]string{"char_code", "name", "value", "nominal"}).
		AddRow(expectedValute.CharCode, expectedValute.Name, expectedValute.Value, expectedValute.Nominal)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT char_code, name, value, nominal FROM currency_rates WHERE char_code = $1 AND date = $2")).
		WithArgs(strings.ToUpper(charCode), date).
		WillReturnRows(rows)

	valute, err := myDB.GetCurrencyRate(ctx, charCode, date)
	require.NoError(t, err)
	require.Equal(t, expectedValute, valute)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestGetCurrencyRate_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	myDB := &DB{db}
	ctx := context.Background()
	charCode := "EUR"
	date := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT char_code, name, value, nominal FROM currency_rates WHERE char_code = $1 AND date = $2")).
		WithArgs(strings.ToUpper(charCode), date).
		WillReturnError(sql.ErrNoRows)

	_, err = myDB.GetCurrencyRate(ctx, charCode, date)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "getCurrencyRate: no valute data for this date"))

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

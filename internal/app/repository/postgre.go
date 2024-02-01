package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vancho-go/GoCurrencyGate/internal/app/models"
)

var (
	ErrNoValuteDataFound = errors.New("no valute data for this date")
)

type DB struct {
	*sql.DB
}

func Initialize(uri string) (*DB, error) {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, fmt.Errorf("initialize: error opening database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("initialize: error verifing database connection: %w", err)
	}

	err = createIfNotExists(db)
	if err != nil {
		return nil, fmt.Errorf("initialize: error creating database structure: %w", err)
	}
	return &DB{db}, nil
}

func createIfNotExists(db *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS currency_rates (
			id SERIAL PRIMARY KEY,
			char_code VARCHAR(3) NOT NULL,
			nominal INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			value NUMERIC NOT NULL,
			date DATE NOT NULL
		);`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("createIfNotExists: %w", err)
	}
	return nil
}

func (db *DB) SaveCurrencyRate(ctx context.Context, valuteCurs *models.ValuteCurs, date time.Time) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("saveCurrencyRate: error beginning transaction: %v", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO currency_rates (char_code, nominal, name, value, date) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return fmt.Errorf("saveCurrencyRate: error preparing insert statement: %v", err)
	}
	defer stmt.Close()

	for _, valute := range valuteCurs.Valutes {
		valute.Value = strings.Replace(valute.Value, ",", ".", -1)
		_, err = stmt.Exec(valute.CharCode, valute.Nominal, valute.Name, valute.Value, date)
		if err != nil {
			tx.Rollback() // Откат в случае ошибки
			return fmt.Errorf("saveCurrencyRate: error executing insert statement: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("saveCurrencyRate: error committing transaction: %v", err)
	}

	return nil
}

func (db *DB) GetCurrencyRate(ctx context.Context, charCode string, date time.Time) (models.APIGetValuteResponse, error) {
	var valute models.APIGetValuteResponse
	row := db.QueryRowContext(ctx, "SELECT char_code, name, value, nominal FROM currency_rates WHERE char_code = $1 AND date = $2",
		strings.ToUpper(charCode), date)
	err := row.Scan(&valute.CharCode, &valute.Name, &valute.Value, &valute.Nominal)
	if errors.Is(err, sql.ErrNoRows) {
		return valute, fmt.Errorf("getCurrencyRate: %w", ErrNoValuteDataFound)
	}
	valute.Date = date.Format("02/01/2006")

	return valute, err
}

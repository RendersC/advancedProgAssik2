package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/crewsrenders/payment-service/internal/domain"
	_ "github.com/lib/pq"
)

// PaymentRepository handles persistence of payment records.
type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, user_id, amount, currency, status, processed_at, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.OrderID, p.UserID, p.Amount, p.Currency, p.Status, p.ProcessedAt, p.Message)
	return err
}

func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `SELECT id, order_id, user_id, amount, currency, status, processed_at, message FROM payments WHERE order_id = $1`
	row := r.db.QueryRowContext(ctx, query, orderID)

	p := &domain.Payment{}
	err := row.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.ProcessedAt, &p.Message)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment for order %s not found", orderID)
	}
	return p, err
}

// GetStats returns aggregated payment statistics via a single SQL query.
func (r *PaymentRepository) GetStats(ctx context.Context) (*domain.PaymentStats, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)                                             AS total_count,
			COUNT(*) FILTER (WHERE status = 'SUCCESS')          AS authorized_count,
			COUNT(*) FILTER (WHERE status = 'FAILED')           AS declined_count,
			COALESCE(SUM(ROUND(amount * 100)::BIGINT), 0)       AS total_amount
		FROM payments`)

	s := &domain.PaymentStats{}
	if err := row.Scan(&s.TotalCount, &s.AuthorizedCount, &s.DeclinedCount, &s.TotalAmount); err != nil {
		return nil, fmt.Errorf("get payment stats: %w", err)
	}
	return s, nil
}

// Migrate creates the payments table if it doesn't exist.
func Migrate(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS payments (
			id           TEXT PRIMARY KEY,
			order_id     TEXT NOT NULL,
			user_id      TEXT NOT NULL,
			amount       DOUBLE PRECISION NOT NULL,
			currency     TEXT NOT NULL,
			status       TEXT NOT NULL,
			processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			message      TEXT
		)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("migrate payments: %w", err)
	}
	return nil
}

func NewPostgresDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

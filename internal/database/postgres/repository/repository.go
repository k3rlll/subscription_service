package repository

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/customerrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

type CreateRequestInput struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   string
	EndDate     string
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

func (r *Repository) CreateSubscription(ctx context.Context, input CreateRequestInput) error {
	query := `insert into subscriptions (id, service_name, price, user_id, start_date, end_date) 
        values ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(
		ctx,
		query,
		input.ID,
		input.ServiceName,
		input.Price,
		input.UserID,
		input.StartDate,
		input.EndDate,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation ex: subs already exists for this user and service
				return fmt.Errorf("subscription already exists: %w", customerrors.ErrAlreadyExists)
			case "23503": // foreign_key_violation
				return fmt.Errorf("user not found: %w", customerrors.ErrNotFound)
			}
		}

		// another errors
		return fmt.Errorf("failed to insert subscription: %w", err)
	}

	return nil
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID) (SubscriptionResponse, error) {
	query := `select id, service_name, price, user_id, start_date, end_date 
		from subscriptions where id = $1`
	row := r.pool.QueryRow(ctx, query, subscriptionID)

	var response SubscriptionResponse

	err := row.Scan(
		&response.ID,
		&response.ServiceName,
		&response.Price,
		&response.UserID,
		&response.StartDate,
		&response.EndDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SubscriptionResponse{}, fmt.Errorf("subscription not found: %w", customerrors.ErrNotFound)
		}
		return SubscriptionResponse{}, fmt.Errorf("failed to query subscription: %w", err)
	}
	return response, nil
}

func (r *Repository) GetListSubs(ctx context.Context, userID uuid.UUID) ([]SubscriptionResponse, error) {

	query := `select id, service_name, price, user_id, start_date, end_date 
        from subscriptions 
        where user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	subs := make([]SubscriptionResponse, 0)

	for rows.Next() {
		var sub SubscriptionResponse
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
		)
		if err != nil {
			return nil, fmt.Errorf("scan subscription row: %w", err)
		}
		subs = append(subs, sub)
	}

	// check fpr error that could occur during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return subs, nil
}

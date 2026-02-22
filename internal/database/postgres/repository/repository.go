package repository

import (
	"context"
	"errors"
	"fmt"
	"main/internal/domain/customerrors"
	domain "main/internal/domain/entity"
	"time"

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

type UpdateSubscriptionRequest struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

func (r *Repository) CreateSubscription(ctx context.Context, input domain.Subscription) error {
	query := `insert into subscriptions 
		(id, service_name, price, user_id, start_date, end_date) 
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

func (r *Repository) GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID) (domain.Subscription, error) {
	query := `select id, service_name, price, user_id, start_date, end_date 
		from subscriptions where id = $1`
	row := r.pool.QueryRow(ctx, query, subscriptionID)

	var response domain.Subscription

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
			return domain.Subscription{}, fmt.Errorf("subscription not found: %w", customerrors.ErrNotFound)
		}
		return domain.Subscription{}, fmt.Errorf("failed to query subscription: %w", err)
	}
	return response, nil
}

func (r *Repository) GetListSubs(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {

	query := `select id, service_name, price, user_id, start_date, end_date 
        from subscriptions 
        where user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	subs := make([]domain.Subscription, 0)

	for rows.Next() {
		var sub domain.Subscription
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

func (r *Repository) UpdateSubscription(ctx context.Context, input domain.Subscription) error {
	query := `update subscriptions 
        set service_name = $1, price = $2, start_date = $3, end_date = $4 
        where id = $5`
	tag, err := r.pool.Exec(
		ctx,
		query,
		input.ServiceName,
		input.Price,
		input.StartDate,
		input.EndDate,
		input.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	// if RowsAffected() returns 0, it means that no subscription with the given ID was found
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription with id %s not found: %w", input.ID, customerrors.ErrNotFound)
	}

	return nil
}

func (r *Repository) DeleteSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	query := `delete from subscriptions where id = $1`
	tag, err := r.pool.Exec(ctx, query, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("subscription with id %s not found: %w", subscriptionID, customerrors.ErrNotFound)
	}
	return nil
}

func (r *Repository) GetOverlappingSubscriptions(ctx context.Context, userID uuid.UUID, serviceName string, start, end time.Time) ([]domain.Subscription, error) {

	query := `
        SELECT id, user_id, service_name, price, start_date, end_date
        FROM subscriptions
        WHERE user_id = $1
          -- optional filter by service name
          AND ($2 = '' OR service_name = $2)
          AND start_date <= $4 
          AND (end_date IS NULL OR end_date >= $3);
    `

	rows, err := r.pool.Query(ctx, query, userID, serviceName, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query overlapping subscriptions: %w", err)
	}
	defer rows.Close()

	// itarete over the rows and scan the data into a slice of domain.Subscription
	response := make([]domain.Subscription, 0)
	for rows.Next() {
		var sub domain.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.ServiceName,
			&sub.Price,
			&sub.StartDate,
			&sub.EndDate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		response = append(response, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return response, nil
}

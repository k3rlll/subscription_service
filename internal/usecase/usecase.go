package usecase

import (
	"context"
	"fmt"
	domain "main/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

//go:generate mockgen -source=usecase.go -destination=./mocks/usecase_mock.go -package=mocks
type Repository interface {
	//CreateSubscription handles the creation of a new subscription based on the provided request payload (user ID, service name, price, start date, end date)
	CreateSubscription(ctx context.Context, input domain.Subscription) error
	//
	//GetSubscriptionByID retrieves a subscription by its unique identifier (UUID) and returns the subscription details or an error if not found
	GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID) (domain.Subscription, error)
	//
	//GetListSubs retrieves a list of subscriptions for a specific user, with support for pagination through limit and offset parameters.
	// It returns a slice of subscriptions or an error if the retrieval fails
	GetListSubs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Subscription, error)
	//
	//UpdateSubscription updates an existing subscription's details based on the provided subscription object,
	// which includes the subscription ID and the new values for the fields to be updated.
	// It returns an error if the update operation fails
	UpdateSubscription(ctx context.Context, input domain.Subscription) error
	//
	//DeleteSubscription deletes a subscription identified by its unique identifier (UUID) and returns an error if the deletion fails
	DeleteSubscription(ctx context.Context, subscriptionID uuid.UUID) error
	//
	//GetOverlappingSubscriptions retrieves a list of subscriptions for a specific user and service that overlap with a given date range (start and end).
	// It returns a slice of overlapping subscriptions or an error if the retrieval fails
	GetOverlappingSubscriptions(ctx context.Context, userID uuid.UUID, serviceName string, start, end time.Time) ([]domain.Subscription, error)
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (u *Usecase) CreateSubscription(ctx context.Context, req domain.Subscription) error {
	subID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate subscription ID: %w", err)
	}
	req.ID = subID
	err = u.repo.CreateSubscription(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (u *Usecase) GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID) (domain.Subscription, error) {
	response, err := u.repo.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("failed to get subscription: %w", err)
	}
	return response, nil
}

func (u *Usecase) GetListSubs(ctx context.Context, req domain.Subscription, limit, offset int) ([]domain.Subscription, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	response, err := u.repo.GetListSubs(ctx, req.UserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of subscriptions: %w", err)
	}
	return response, nil
}

func (u *Usecase) UpdateSubscription(ctx context.Context, req domain.Subscription) error {
	err := u.repo.UpdateSubscription(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil

}

func (u *Usecase) DeleteSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	err := u.repo.DeleteSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

// TODO: разберись с подсчетом стоимости
func (u *Usecase) CalculateTotalCost(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	reqStart time.Time,
	reqEnd time.Time,
) (int64, error) {

	// prevent calculating cost for the too far future
	// if a request will contain range from 01-2026 to 12-2099,
	// the cost will be too huge
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	if reqEnd.After(currentMonth) {
		reqEnd = currentMonth
	}

	subs, err := u.repo.GetOverlappingSubscriptions(ctx, userID, serviceName, reqStart, reqEnd)
	if err != nil {
		return 0, err
	}

	var totalCost int64 = 0

	for _, sub := range subs {
		calcStart := sub.StartDate
		if reqStart.After(calcStart) {
			calcStart = reqStart
		}

		calcEnd := reqEnd

		if sub.EndDate != nil && sub.EndDate.Before(reqEnd) {
			calcEnd = *sub.EndDate
		}

		// protect against invalid date ranges, where the calculated start is after the calculated end
		// but actually it is not possible, because handler already checks the validity of the date range,
		// however, it will not be overkill to add this check, just in case
		if calcStart.After(calcEnd) {
			continue
		}

		// calculate the number of months between calcStart and calcEnd
		// formula:
		// (year2 - year1)*12
		// + (month2 - month1)
		// + 1 (including the month of start date)
		months := int(calcEnd.Year()-calcStart.Year())*12 + int(calcEnd.Month()-calcStart.Month()) + 1

		totalCost += int64(months) * sub.Price
	}

	return totalCost, nil
}

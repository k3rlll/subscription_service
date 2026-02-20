package usecase

import (
	"context"
	"fmt"
	"main/internal/domain/customerrors"

	"github.com/google/uuid"
)

type Repository interface {
	CreateSubscription(ctx context.Context, input CreateRequestOutput) error
	GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID) (SubscriptionResponse, error)
	GetListSubs(ctx context.Context, userID uuid.UUID) ([]SubscriptionResponse, error)
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

type CreateRequestInput struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     string
}

type CreateRequestOutput struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
}

type ListSubscriptionsRequest struct {
	UserID string `json:"user_id"`
	Limit  string `json:"limit"`
	Offset string `json:"offset"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

func (u *Usecase) CreateSubscription(ctx context.Context, input CreateRequestInput) error {
	var req CreateRequestOutput
	subID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate subscription ID: %w", err)
	}

	parsedUserID, err := uuid.Parse(input.UserID)
	if err != nil {
		return fmt.Errorf("failed to parse UUID: %w", customerrors.ErrInvalidRequest)
	}
	req.ID = subID
	req.ServiceName = input.ServiceName
	req.Price = input.Price
	req.UserID = parsedUserID
	req.StartDate = input.StartDate
	req.EndDate = input.EndDate
	err = u.repo.CreateSubscription(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (u *Usecase) GetSubscriptionByID(ctx context.Context, subscriptionID string) (SubscriptionResponse, error) {
	subsUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		return SubscriptionResponse{}, fmt.Errorf("failed to parse subscription ID: %w", customerrors.ErrInvalidRequest)
	}
	response, err := u.repo.GetSubscriptionByID(ctx, subsUUID)
	if err != nil {
		return SubscriptionResponse{}, fmt.Errorf("failed to get subscription: %w", err)
	}
	return response, nil
}

func (u *Usecase) GetListSubs(ctx context.Context, req ListSubscriptionsRequest) ([]SubscriptionResponse, error) {
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", customerrors.ErrInvalidRequest)
	}

	response, err := u.repo.GetListSubs(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of subscriptions: %w", err)
	}
	return response, nil
}

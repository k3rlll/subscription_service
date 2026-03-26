package mocks

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "main/internal/domain/entity"
	uc "main/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupTest(t *testing.T) (*uc.Usecase, *MockRepository) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	uc := uc.NewUsecase(repo)
	return uc, repo
}

func TestUsecase_CreateSubscription(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := domain.Subscription{
			UserID:      uuid.New(),
			ServiceName: "Netflix",
			Price:       1500,
		}

		repo.EXPECT().CreateSubscription(ctx, gomock.Any()).Return(nil).Times(1)

		err := uc.CreateSubscription(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		repo.EXPECT().CreateSubscription(ctx, gomock.Any()).Return(errors.New("db error")).Times(1)

		err := uc.CreateSubscription(ctx, domain.Subscription{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create subscription")
	})
}

func TestUsecase_GetSubscriptionByID(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		subID := uuid.New()
		expectedSub := domain.Subscription{ID: subID, ServiceName: "Spotify"}

		repo.EXPECT().GetSubscriptionByID(ctx, subID).Return(expectedSub, nil).Times(1)

		sub, err := uc.GetSubscriptionByID(ctx, subID)
		assert.NoError(t, err)
		assert.Equal(t, expectedSub, sub)
	})
}

func TestUsecase_GetListSubs(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	req := domain.Subscription{UserID: userID}

	t.Run("success with default pagination", func(t *testing.T) {
		invalidLimit, invalidOffset := 0, -5
		expectedLimit, expectedOffset := 10, 0

		repo.EXPECT().GetListSubs(ctx, userID, expectedLimit, expectedOffset).Return([]domain.Subscription{}, nil).Times(1)

		subs, err := uc.GetListSubs(ctx, req, invalidLimit, invalidOffset)
		assert.NoError(t, err)
		assert.NotNil(t, subs)
	})

	t.Run("success with custom pagination", func(t *testing.T) {
		limit, offset := 50, 20

		repo.EXPECT().GetListSubs(ctx, userID, limit, offset).Return([]domain.Subscription{}, nil).Times(1)

		_, err := uc.GetListSubs(ctx, req, limit, offset)
		assert.NoError(t, err)
	})
}

func TestUsecase_UpdateSubscription(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		req := domain.Subscription{ID: uuid.New(), ServiceName: "Updated"}

		repo.EXPECT().UpdateSubscription(ctx, req).Return(nil).Times(1)

		err := uc.UpdateSubscription(ctx, req)
		assert.NoError(t, err)
	})
}

func TestUsecase_DeleteSubscription(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		subID := uuid.New()

		repo.EXPECT().DeleteSubscription(ctx, subID).Return(nil).Times(1)

		err := uc.DeleteSubscription(ctx, subID)
		assert.NoError(t, err)
	})
}

func TestUsecase_CalculateTotalCost(t *testing.T) {
	uc, repo := setupTest(t)
	ctx := context.Background()
	userID := uuid.New()
	serviceName := "Yandex Plus"

	t.Run("success calculation - past dates", func(t *testing.T) {
		reqStart := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		reqEnd := time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)

		mockSubs := []domain.Subscription{
			{
				StartDate: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   nil,
				Price:     299,
			},
		}

		repo.EXPECT().GetOverlappingSubscriptions(ctx, userID, serviceName, reqStart, reqEnd).Return(mockSubs, nil).Times(1)

		cost, err := uc.CalculateTotalCost(ctx, userID, serviceName, reqStart, reqEnd)
		assert.NoError(t, err)
		assert.Equal(t, int64(897), cost)
	})

	t.Run("success calculation - future truncation", func(t *testing.T) {
		reqStart := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		reqEnd := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)

		now := time.Now()
		expectedEnd := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

		mockSubs := []domain.Subscription{}

		repo.EXPECT().GetOverlappingSubscriptions(ctx, userID, serviceName, reqStart, expectedEnd).Return(mockSubs, nil).Times(1)

		_, err := uc.CalculateTotalCost(ctx, userID, serviceName, reqStart, reqEnd)
		assert.NoError(t, err)
	})

	t.Run("success calculation - specific end date", func(t *testing.T) {
		reqStart := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		reqEnd := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)

		subEnd := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)

		mockSubs := []domain.Subscription{
			{
				StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   &subEnd,
				Price:     100,
			},
		}

		repo.EXPECT().GetOverlappingSubscriptions(ctx, userID, serviceName, reqStart, reqEnd).Return(mockSubs, nil).Times(1)

		cost, err := uc.CalculateTotalCost(ctx, userID, serviceName, reqStart, reqEnd)
		assert.NoError(t, err)
		assert.Equal(t, int64(200), cost)
	})
}

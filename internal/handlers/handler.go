package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"main/internal/domain/customerrors"
	errDTO "main/pkg/errDTO"

	"github.com/go-playground/validator/v10"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	e       *echo.Echo
	usecase AuthUsecase
	logger  *slog.Logger
}

func NewHandler(e *echo.Echo, usecase AuthUsecase, logger *slog.Logger) *Handler {
	return &Handler{
		e:       e,
		usecase: usecase,
		logger:  logger,
	}
}

type AuthUsecase interface {
	// CreateSubscription handles the creation of a new subscription based on the incoming request
	CreateSubscription(c echo.Context, req CreateRequest) error

	// GetCalculations handles the retrieval of calculations based on the provided query parameters
	GetCalculations(c echo.Context, req GetCalculationsRequest) (GetCalculationsResponse, error)

	// GetSubscription handles the retrieval of subscription details based on the provided subscription ID
	GetSubscriptionByID(c echo.Context, subscriptionID string) (SubscriptionResponse, error)

	// GetListSubs handles the retrieval of a list of subscriptions for a specific user based on the provided user ID
	GetListSubs(c echo.Context, req ListSubscriptionsRequest) ([]SubscriptionResponse, error)

	// UpdateSubscription handles the update of an existing subscription
	// based on the provided subscription ID and request payload
	UpdateSubscription(c echo.Context, req UpdateRequest) error

	// DeleteSubscription handles the deletion of a subscription based on the provided subscription ID
	DeleteSubscription(c echo.Context, subscriptionID string) error
}

// Data transfer object for creating a subscription (DTO)
type CreateRequest struct {
	/*“service_name”: “Yandex Plus”,
	  “price”: 400,
	  “user_id”: “60601fee-2bf1-4721-ae6f-7636e79a0cba”,
	  “start_date”: “07-2025”*/

	ServiceName string `json:"service_name" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	UserID      string `json:"user_id" validate:"required,uuid4"`
	StartDate   string `json:"start_date" validate:"required"`
	EndDate     string `json:"end_date" validate:"required"`
}

type SubscriptionResponse struct {
	ID          string `json:"id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type UpdateRequest struct {
	ID          string `json:"id" validate:"required,uuid4"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type ListSubscriptionsRequest struct {
	UserID string `query:"user_id" validate:"required,uuid4"`
	Limit  string `query:"limit" validate:"required,min=1,max=100"`
	Offset string `query:"offset" validate:"required, min=0"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

type GetCalculationsRequest struct {
	UserID      string `query:"user_id" validate:"required,uuid4"`
	ServiceName string `query:"service_name"`
	StartDate   string `query:"start_date"`
	EndDate     string `query:"end_date"`
}

type GetCalculationsResponse struct {
	TotalCost int `json:"total_cost"`
}

// POST
// CreateSubscription handles the creation of a new subscription based on the incoming request
func (h *Handler) CreateSubscription(c echo.Context) error {
	var req CreateRequest

	// actualy could be more specific and check if the error is a binding error,
	// there is opportunity to create custom bind funtion with more specific error messages
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind create subscription request", "error", err)
		return errDTO.New(http.StatusBadRequest, "invalid_request", err.Error())
	}

	// validation via github.com/go-playground/validator
	if err := c.Validate(&req); err != nil {
		// check wether the error is a validation error
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			h.logger.Info("validation failed", "error", validationErrs.Error())
			return errDTO.New(http.StatusBadRequest, "validation_error", validationErrs.Error())
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return fmt.Errorf("validation failed: %w", err)
	}

	// check if start date is after end date
	if req.StartDate > req.EndDate {
		return errDTO.New(http.StatusBadRequest, "invalid_date_range", "start date cannot be after end date")
	}

	err := h.usecase.CreateSubscription(c, req)
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrInvalidRequest):
			h.logger.Info("invalid request", "error", err)
			return errDTO.New(http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, customerrors.ErrNotFound):
			h.logger.Info("resource not found", "error", err)
			return errDTO.New(http.StatusNotFound, "not_found", "Resource not found")
		case errors.Is(err, customerrors.ErrAlreadyExists):
			h.logger.Info("resource already exists", "error", err)
			return errDTO.New(http.StatusConflict, "already_exists", "Resource already exists")
		default:
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription created",
	})
}

// GET

// GetCalculations handles the retrieval of calculations based on the provided query parameters
func (h *Handler) GetCalculations(c echo.Context) error {

	// there is c.Param in echo framework, but it is used for path parameters, for example: /calculations/:id
	// in REST i should use query parameters,
	// for instance: /calculations?user_id=blabla&service_name==blabla&start_date=blabla&end_date=blabla
	// so i will use c.QueryParam to get query parameters from the request
	var req GetCalculationsRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Info("failed to bind query params", "error", err)
		return errDTO.New(http.StatusBadRequest, "invalid_request", "invalid query parameters")
	}

	if err := c.Validate(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			h.logger.Info("validation failed", "error", validationErrs.Error())
			return errDTO.New(http.StatusBadRequest, "validation_error", validationErrs.Error())
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return fmt.Errorf("validation failed: %w", err)
	}

	response, err := h.usecase.GetCalculations(c, req)
	if err != nil {
		return errDTO.New(http.StatusInternalServerError, "internal_server_error", err.Error())
	}
	h.logger.Info("Calculations retrieved successfully", "user_id", req.UserID, "service_name", req.ServiceName)
	return c.JSON(http.StatusOK, response)
}

// GetSubscription handles the retrieval of subscription details based on the provided subscription ID
func (h *Handler) GetSubscriptionByID(c echo.Context) error {
	var subscriptionID string
	if err := c.Bind(&subscriptionID); err != nil {
		h.logger.Info("Failed to bind get subscription request", "error", err)
		return errDTO.New(http.StatusBadRequest, "invalid_request", "Invalid request payload")
	}

	if err := c.Validate(&subscriptionID); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			h.logger.Info("validation failed", "error", validationErrs.Error())
			return errDTO.New(http.StatusBadRequest, "validation_error", validationErrs.Error())
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return fmt.Errorf("validation failed: %w", err)
	}

	response, err := h.usecase.GetSubscriptionByID(c, subscriptionID)
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrInvalidRequest):
			h.logger.Info("invalid request", "error", err)
			return errDTO.New(http.StatusBadRequest, "invalid_request", err.Error())
		case errors.Is(err, customerrors.ErrNotFound):
			h.logger.Info("subscription not found", "error", err)
			return errDTO.New(http.StatusNotFound, "not_found", "Subscription not found")
		default:
			return fmt.Errorf("failed to get subscription: %w", err)
		}

	}
	return c.JSON(http.StatusOK, response)
}

// GetListSubs handles the retrieval of a list of subscriptions for a specific user
// using pagination parameters (limit and offset)
// query example: user_id=123&limit=10&offset=0
func (h *Handler) GetListSubs(c echo.Context) error {
	var listReq ListSubscriptionsRequest
	if err := c.Bind(&listReq); err != nil {
		h.logger.Info("Failed to bind list subscriptions request", "error", err)
		return errDTO.New(http.StatusBadRequest, "invalid_request", "Invalid request payload")
	}

	if err := c.Validate(&listReq); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			h.logger.Info("validation failed", "error", validationErrs.Error())
			return errDTO.New(http.StatusBadRequest, "validation_error", validationErrs.Error())
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return fmt.Errorf("validation failed: %w", err)
	}

	response, err := h.usecase.GetListSubs(c, listReq)
	if err != nil {
		return fmt.Errorf("Failed to get list of subscriptions: %w", err)
	}

	h.logger.Info("List of subscriptions retrieved successfully", "user_id", listReq.UserID)
	return c.JSON(http.StatusOK, response)
}

// PUT
// UpdateSubscription handles the update of an existing subscription
// based on the provided subscription ID and request payload
func (h *Handler) UpdateSubscription(c echo.Context) error {
	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind update request", "error", err)
		return errDTO.New(http.StatusBadRequest, "invalid_request", "Invalid request payload")
	}

	err := h.usecase.UpdateSubscription(c, req)
	if err != nil {
		h.logger.Error("Failed to update subscription", "error", err)
		return errDTO.New(http.StatusInternalServerError, "internal_server_error", "Failed to update subscription")
	}
	h.logger.Info("Subscription updated successfully", "subscription_id", req.ID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription updated",
	})
}

// DELETE
// DeleteSubscription handles the deletion of a subscription based on the provided subscription ID
func (h *Handler) DeleteSubscription(c echo.Context) error {
	var subscriptionID string
	if err := c.Bind(&subscriptionID); err != nil {
		h.logger.Error("Failed to bind delete subscription request", "error", err)
		return errDTO.New(http.StatusBadRequest, "invalid_request", "Invalid request payload")
	}
	if err := c.Validate(subscriptionID); err != nil {
		h.logger.Error("Failed to validate subscription ID", "error", err)
		return errDTO.New(http.StatusBadRequest, "invalid_request", "Invalid subscription ID")
	}
	err := h.usecase.DeleteSubscription(c, subscriptionID)
	if err != nil {
		h.logger.Error("Failed to delete subscription", "error", err)
		return errDTO.New(http.StatusInternalServerError, "internal_server_error", "Failed to delete subscription")
	}
	h.logger.Info("Subscription deleted successfully", "subscription_id", subscriptionID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription deleted",
	})
}

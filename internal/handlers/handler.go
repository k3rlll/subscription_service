package handlers

import (
	"log/slog"
	"net/http"

	errDTO "main/pkg/errDTO"

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
	GetSubscription(c echo.Context, subscriptionID string) (SubscriptionResponse, error)

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
	UserID string `json:"user_id" validate:"required,uuid4"`
	Limit  string `json:"limit"`
	Offset string `json:"offset"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

type GetCalculationsRequest struct {
	UserID      string `json:"user_id" validate:"required,uuid4"`
	ServiceName string `json:"service_name"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type GetCalculationsResponse struct {
	TotalCost int `json:"total_cost"`
}

// POST
// CreateSubscription handles the creation of a new subscription based on the incoming request
func (h *Handler) CreateSubscription(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, errDTO.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
		})
	}

	err := h.usecase.CreateSubscription(c, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errDTO.ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to create subscription",
		})
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

	req.UserID = c.QueryParam("user_id")
	req.ServiceName = c.QueryParam("service_name")
	req.StartDate = c.QueryParam("start_date")
	req.EndDate = c.QueryParam("end_date")

	if req.UserID == "" || req.UserID == "null" || req.ServiceName == "" || req.ServiceName == "null" || req.StartDate == "" || req.StartDate == "null" || req.EndDate == "" || req.EndDate == "null" {
		h.logger.Error("Missing required query parameters", "user_id", req.UserID, "service_name", req.ServiceName, "start_date", req.StartDate, "end_date", req.EndDate)
		return c.JSON(http.StatusBadRequest, errDTO.ErrorResponse{
			Error:   "invalid_request",
			Message: "All query parameters (user_id, service_name, start_date, end_date) are required",
		})
	}

	response, err := h.usecase.GetCalculations(c, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errDTO.ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve calculations",
		})
	}
	h.logger.Info("Calculations retrieved successfully", "user_id", req.UserID, "service_name", req.ServiceName)
	return c.JSON(http.StatusOK, response)
}

// GetSubscription handles the retrieval of subscription details based on the provided subscription ID
func (h *Handler) GetSubscription(c echo.Context) error {
	subscriptionID := c.QueryParam("id")
	if subscriptionID == "" || subscriptionID == "null" {
		h.logger.Error("Missing required query parameter: id", "id", subscriptionID)
		return c.JSON(http.StatusBadRequest, errDTO.ErrorResponse{
			Error:   "invalid_request",
			Message: "Subscription ID is required",
		})
	}
	response, err := h.usecase.GetSubscription(c, subscriptionID)
	if err != nil {
		h.logger.Error("Failed to retrieve subscription", "error", err)
		return c.JSON(http.StatusInternalServerError, errDTO.ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve subscription",
		})
	}
	h.logger.Info("Subscription retrieved successfully", "subscription_id", subscriptionID)
	return c.JSON(http.StatusOK, response)
}

// GetListSubs handles the retrieval of a list of subscriptions for a specific user
// using pagination parameters (limit and offset)
// query example: user_id=123&limit=10&offset=0
func (h *Handler) GetListSubs(c echo.Context) error {
	userID := c.QueryParam("user_id")
	limit := c.QueryParam("limit")
	offset := c.QueryParam("offset")

	if userID == "" || limit == "" || offset == "" ||
		userID == "null" || limit == "null" || offset == "null" {
		h.logger.Error("Missing required query parameters", "user_id", userID, "limit", limit, "offset", offset)
		return c.JSON(http.StatusBadRequest, errDTO.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID, limit and offset are required",
		})
	}

	response, err := h.usecase.GetListSubs(c, ListSubscriptionsRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("Failed to retrieve list of subscriptions", "error", err)
		return c.JSON(http.StatusInternalServerError, errDTO.ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to retrieve list of subscriptions",
		})
	}
	h.logger.Info("List of subscriptions retrieved successfully", "user_id", userID)
	return c.JSON(http.StatusOK, response)
}

// PUT
// UpdateSubscription handles the update of an existing subscription
// based on the provided subscription ID and request payload
func (h *Handler) UpdateSubscription(c echo.Context) error {
	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind update request", "error", err)
		return c.JSON(http.StatusBadRequest, errDTO.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
		})
	}

	err := h.usecase.UpdateSubscription(c, req)
	if err != nil {
		h.logger.Error("Failed to update subscription", "error", err)
		return c.JSON(http.StatusInternalServerError, errDTO.ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to update subscription",
		})
	}
	h.logger.Info("Subscription updated successfully", "subscription_id", req.ID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription updated",
	})
}

// DELETE
// DeleteSubscription handles the deletion of a subscription based on the provided subscription ID
func (h *Handler) DeleteSubscription(c echo.Context) error {
	subscriptionID := c.QueryParam("id")
	if subscriptionID == "" || subscriptionID == "null" {
		h.logger.Error("Missing required query parameter: id", "id", subscriptionID)
		return c.JSON(http.StatusBadRequest, errDTO.ErrorResponse{
			Error:   "invalid_request",
			Message: "Subscription ID is required",
		})
	}
	err := h.usecase.DeleteSubscription(c, subscriptionID)
	if err != nil {
		h.logger.Error("Failed to delete subscription", "error", err)
		return c.JSON(http.StatusInternalServerError, errDTO.ErrorResponse{
			Error:   "internal_server_error",
			Message: "Failed to delete subscription",
		})
	}
	h.logger.Info("Subscription deleted successfully", "subscription_id", subscriptionID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription deleted",
	})
}

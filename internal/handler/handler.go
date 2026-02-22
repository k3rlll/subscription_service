package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"main/internal/domain/customerrors"
	domain "main/internal/domain/entity"

	utils "main/pkg/utils"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

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

// AuthUsecase defines the interface for the use case layer that the handler will interact with
type AuthUsecase interface {
	// CreateSubscription handles the creation of a new subscription based on the incoming request
	CreateSubscription(ctx context.Context, req domain.Subscription) error

	// CalculateTotalCost calculates the total cost of subscriptions for a user
	CalculateTotalCost(ctx context.Context,
		userID uuid.UUID,
		serviceName string,
		reqStart time.Time,
		reqEnd time.Time) (int64, error)

	// GetSubscription handles the retrieval of subscription details based on the provided subscription ID
	GetSubscriptionByID(ctx context.Context, subscriptionID uuid.UUID) (domain.Subscription, error)

	// GetListSubs handles the retrieval of a list of subscriptions for a specific user based on the provided user ID
	GetListSubs(ctx context.Context, req domain.Subscription) ([]domain.Subscription, error)

	// UpdateSubscription handles the update of an existing subscription
	// based on the provided subscription ID and request payload
	UpdateSubscription(ctx context.Context, req domain.Subscription) error

	// DeleteSubscription handles the deletion of a subscription based on the provided subscription ID
	DeleteSubscription(ctx context.Context, subscriptionID uuid.UUID) error
}

// Data transfer object for creating a subscription (DTO)
type CreateRequest struct {
	ServiceName string `json:"service_name" validate:"required"`
	Price       int64  `json:"price" validate:"required,gte=0"`
	UserID      string `json:"user_id" validate:"required,uuid4"`
	StartDate   string `json:"start_date" validate:"required,datetime=01-2006"`
	EndDate     string `json:"end_date" validate:"omitempty,datetime=01-2006"`
}

type CreateInput struct {
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

type SubscriptionResponse struct {
	ID          string     `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

type UpdateRequest struct {
	ID          string `json:"id" validate:"required,uuid4"`
	ServiceName string `json:"service_name" validate:"required"`
	Price       int64  `json:"price" validate:"required,gte=0"`
	StartDate   string `json:"start_date" validate:"required,datetime=01-2006"`
	EndDate     string `json:"end_date" validate:"omitempty,datetime=01-2006"`
}

type UpdateInput struct {
	ID          string     `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

type ListSubscriptionsRequest struct {
	UserID string `query:"user_id" validate:"required,uuid4"`
	Limit  int    `query:"limit" validate:"required,min=1,max=100"`
	Offset int    `query:"offset" validate:"min=0"`
}

type GetCalculationsRequest struct {
	UserID      string `query:"user_id" validate:"required,uuid4"`
	ServiceName string `query:"service_name" validate:"omitempty"`
	StartDate   string `query:"start_date" validate:"required,datetime=01-2006"`
	EndDate     string `query:"end_date" validate:"omitempty,datetime=01-2006"`
}

type GetCalculationsInput struct {
	UserID      uuid.UUID
	ServiceName string
	StartDate   time.Time
	EndDate     *time.Time
}

type GetCalculationsResponse struct {
	TotalCost int `json:"total_cost"`
}

type DeleteSubscriptionRequest struct {
	ID string `json:"id" validate:"required,uuid4"`
}

// POST
// CreateSubscription handles the creation of a new subscription based on the incoming request
func (h *Handler) CreateSubscription(c echo.Context) error {
	var req CreateRequest

	// actualy could be more specific and check if the error is a binding error,
	// there is opportunity to create custom bind funtion with more specific error messages
	if err := c.Bind(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return echo.NewHTTPError(http.StatusBadRequest, "request body is empty")
		}
		return err
	}
	//

	// validation via github.com/go-playground/validator
	if err := c.Validate(&req); err != nil {
		// check wether the error is a validation error
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return err
	}
	//
	//
	//

	//
	uuidUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid_user_id: Invalid user_id format")
	}
	//
	//
	//

	//
	// parse start and end dates to time.Time
	startTime, endTimePtr, err := utils.ParseDate(req.StartDate, req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := domain.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      uuidUserID,
		StartDate:   startTime,
		EndDate:     endTimePtr,
	}

	err = h.usecase.CreateSubscription(c.Request().Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrInvalidRequest):
			return echo.NewHTTPError(http.StatusBadRequest, "invalid_request: "+err.Error())
		case errors.Is(err, customerrors.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "not_found: Resource not found")
		case errors.Is(err, customerrors.ErrAlreadyExists):
			return echo.NewHTTPError(http.StatusConflict, "already_exists: Resource already exists")
		default:
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	// if subscription is created successfully, we return 201 Created status code with a message
	// but actually it would be better to return the created subscription with its ID, for example:
	// return c.JSON(http.StatusCreated, map[string]interface{}{
	// 	"message": "Subscription created",
	// 	"subscription": createdSubscription,
	// })
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription created",
	})
}

// GET

// GetCalculations return total cost of subscriptions for a user
// based on the provided query parameters
// (user_id, service_name, start_date, end_date)
// service_name can be empty,
// then we will calculate total cost for all services of the user
// for instance: /calculations?user_id=blabla&service_name==blabla&start_date=blabla&end_date=blabla
func (h *Handler) GetCalculations(c echo.Context) error {

	// there is c.Param in echo framework, but it is used for path parameters, for example: /calculations/:id
	// in REST i should use query parameters,
	// for instance: /calculations?user_id=blabla&service_name==blabla&start_date=blabla&end_date=blabla
	// so i will use c.QueryParam to get query parameters from the request
	var req GetCalculationsRequest

	// bind query parameters to the struct
	err := echo.QueryParamsBinder(c).
		String("user_id", &req.UserID).
		String("service_name", &req.ServiceName).
		String("start_date", &req.StartDate).
		String("end_date", &req.EndDate).
		BindError()

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}
	//

	//
	if err := c.Validate(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return fmt.Errorf("validation failed: %w", err)
	}

	startTime, endTimePtr, err := utils.ParseDate(req.StartDate, req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	//
	//
	uuidUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return err
	}

	//if end date is not provided, we will use current date as end date for calculation
	var endTime time.Time
	if endTimePtr != nil {
		endTime = *endTimePtr
	} else {
		endTime = time.Now()
	}

	response, err := h.usecase.CalculateTotalCost(c.Request().Context(), uuidUserID, req.ServiceName, startTime, endTime)
	if err != nil {
		if errors.Is(err, customerrors.ErrInvalidRequest) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid_request: "+err.Error())
		}
		return fmt.Errorf("failed to calculate total cost: %w", err)
	}
	// or just total cost without message, for instance: return c.JSON(http.StatusOK, map[string]interface{}{
	// 	"total_cost": response,
	// })
	return c.JSON(http.StatusOK, fmt.Sprintf("%d рублей", response))
}

// GetSubscription handles the retrieval of subscription details based on the provided subscription ID
// for instance: api/v1/subscriptions/93074b79-9c8c-4b60-82df-0bdd0aaa2b04
func (h *Handler) GetSubscriptionByID(c echo.Context) error {
	var subscriptionID string

	// bind query parameters to the struct
	subscriptionID = c.Param("id")
	//

	//

	subscriptionUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid_subscription_id: Invalid subscription ID format")
	}

	response, err := h.usecase.GetSubscriptionByID(c.Request().Context(), subscriptionUUID)
	if err != nil {
		switch {
		case errors.Is(err, customerrors.ErrInvalidRequest):
			return echo.NewHTTPError(http.StatusBadRequest, "invalid_request: "+err.Error())
		case errors.Is(err, customerrors.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "not_found: Subscription not found")
		default:
			return fmt.Errorf("failed to get subscription: %w", err)
		}

	}
	//
	//
	return c.JSON(http.StatusOK, response)
}

// GetListSubs handles the retrieval of a list of subscriptions for a specific user
// using pagination parameters (limit and offset)
// query example: user_id=123&limit=10&offset=0
func (h *Handler) GetListSubs(c echo.Context) error {
	var listReq ListSubscriptionsRequest

	err := echo.QueryParamsBinder(c).
		String("user_id", &listReq.UserID).
		Int("limit", &listReq.Limit).
		Int("offset", &listReq.Offset).
		BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid query parameters")
	}
	//

	//
	if err := c.Validate(&listReq); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return fmt.Errorf("validation failed: %w", err)
	}
	//
	uuidUserID, err := uuid.Parse(listReq.UserID)
	if err != nil {
		return err
	}

	response, err := h.usecase.GetListSubs(c.Request().Context(), domain.Subscription{UserID: uuidUserID})
	if err != nil {
		return fmt.Errorf("Failed to get list of subscriptions: %w", err)
	}

	//
	//if len(response) == 0 {
	//	return echo.NewHTTPError(http.StatusNotFound, "not_found: No subscriptions found for the user")
	//}
	//
	return c.JSON(http.StatusOK, response)
}

// PUT
// UpdateSubscription handles the update of an existing subscription
// based on the provided subscription ID and request payload
// for instance: api/v1/subscriptions/93074b79-9c8c-4b60-82df-0bdd0aaa2b04
func (h *Handler) UpdateSubscription(c echo.Context) error {
	var req UpdateRequest
	req.ID = c.Param("id") // get subscription ID from path parameter

	// bind request body to the struct
	if err := c.Bind(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return echo.NewHTTPError(http.StatusBadRequest, "request body is empty")
		}
		return err
	}
	//
	//

	//
	if err := c.Validate(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	//

	// parse start and end dates to time.Time
	startTime, endTimePtr, err := utils.ParseDate(req.StartDate, req.EndDate)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid_date: ", err)
	}

	uuidSubID, err := uuid.Parse(req.ID)
	if err != nil {
		return err
	}

	var input domain.Subscription
	input.ID = uuidSubID
	input.ServiceName = req.ServiceName
	input.Price = req.Price
	input.StartDate = startTime
	input.EndDate = endTimePtr

	err = h.usecase.UpdateSubscription(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "not_found: Subscription not found")
		}
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription updated",
	})
}

// DELETE
// DeleteSubscription handles the deletion of a subscription based on the provided subscription ID
// for instance: api/v1/subscriptions/93074b79-9c8c-4b60-82df-0bdd0aaa2b04
func (h *Handler) DeleteSubscription(c echo.Context) error {
	var req DeleteSubscriptionRequest
	req.ID = c.Param("id")
	//
	//

	//
	if err := c.Validate(&req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return echo.NewHTTPError(http.StatusBadRequest, formatValidationError(err))
		}
		// 500 Internal Server Error: mistake of the developer or the validation package
		return fmt.Errorf("validation failed: %w", err)
	}

	uuidSubID, err := uuid.Parse(req.ID)
	if err != nil {
		return err
	}

	//
	//
	//

	err = h.usecase.DeleteSubscription(c.Request().Context(), uuidSubID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "not_found: Subscription not found")
		}
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Subscription deleted",
	})
}

// formatValidationError is a helper function that formats validation errors into a map of field names to error messages
func formatValidationError(err error) map[string]string {
	erro := make(map[string]string)
	var vErrs validator.ValidationErrors

	if errors.As(err, &vErrs) {
		for _, f := range vErrs {
			// f.Field() — name, f.Tag() — violated rule (required, email, etc.)
			erro[f.Field()] = fmt.Sprintf("failed on the '%s' tag", f.Tag())
		}
	} else {
		erro["error"] = err.Error()
	}
	return erro
}

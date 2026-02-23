package app

import (
	handler "main/internal/handler"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// MapRoutes maps the API routes to their corresponding handler functions.
func MapRoutes(
	e *echo.Echo,
	userHandler *handler.Handler,
) {
	// Swagger documentation route
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	v1 := e.Group("/api/v1")
	{
		v1.POST("/subscriptions", userHandler.CreateSubscription)
		v1.GET("/subscriptions", userHandler.GetListSubs)

		v1.GET("/subscriptions/calculations", userHandler.GetCalculations)

		v1.GET("/subscriptions/:id", userHandler.GetSubscriptionByID)
		v1.PUT("/subscriptions/:id", userHandler.UpdateSubscription)
		v1.DELETE("/subscriptions/:id", userHandler.DeleteSubscription)
	}
}

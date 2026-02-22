package app

import (
	handler "main/internal/handler"

	"github.com/labstack/echo/v4"
)

func MapRoutes(
	e *echo.Echo,
	userHandler *handler.Handler,
) {
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

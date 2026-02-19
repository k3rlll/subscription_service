package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func MapRoutes(
	e *echo.Echo,
	authHandler *handler.AuthHandler,
	authUsecase AuthUsecase,
	logger *slog.Logger,
) {
	// Middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper:   e.DefaultSkipper,
		LogURI:    true,
		LogMethod: true,
		LogStatus: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {

			if v.Error != nil {
				logger.Error("HTTP request error",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"error", v.Error,
				)
				return nil
			}

			logger.Info("HTTP request",
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"error", v.Error,
			)

			return nil
		},
	},
	))

	//routes

	logger.Info("HTTP routes mapped successfully")
}

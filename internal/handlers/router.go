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
	// Custom request logger middleware to log HTTP requests with structured logging
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:     true,
		LogMethod:  true,
		LogStatus:  true,
		LogError:   true,
		LogLatency: true, // duration of request processing for monitoring performance
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {

			// gather all log fields in a structured way
			args := []any{
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency.String(),
			}

			if v.Error != nil {
				args = append(args, "error", v.Error.Error())
			}

			// 500 and above are server errors, which we log as ERROR level
			if v.Status >= 500 {
				logger.Error("HTTP server error", args...)
			} else if v.Status >= 400 {
				// 400-499 are client errors, which we log as WARN level
				logger.Warn("HTTP client error", args...)
			} else {
				// 200-399 are successful responses, which we log as INFO level
				logger.Info("HTTP request success", args...)
			}

			return nil
		},
	}))
}

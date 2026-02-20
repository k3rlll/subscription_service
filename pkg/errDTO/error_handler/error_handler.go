package errorhandler

import (
	"fmt"
	"log/slog"
	errDTO "main/pkg/errDTO"
	"net/http"

	"github.com/labstack/echo/v4"
)

// proccesses all errors returned by handlers in a consistent way
func CustomHTTPErrorHandler(err error, c echo.Context, logger *slog.Logger) {
	if ErrDTO, ok := err.(*errDTO.ErrorDTO); ok {
		// if its error with no sensitive data, we can just return it as is
		c.JSON(ErrDTO.HTTPStatus, ErrDTO)
		return
	}

	// if it's an Echo HTTP error, we can extract the status code and message
	// and return it in a consistent format to the client
	if echoErr, ok := err.(*echo.HTTPError); ok {
		c.JSON(echoErr.Code, errDTO.ErrorDTO{
			Code:    "router_error",
			Message: fmt.Sprintf("%v", echoErr.Message),
		})
		return
	}

	//if it's some other error, we log it as an unhandled error and return a generic message to the client

	logger.Error("Unhandled API Error", "error", err.Error(), "path", c.Path())

	c.JSON(http.StatusInternalServerError, errDTO.ErrorDTO{
		Code:    "internal_server_error",
		Message: "something went wrong",
	})
}

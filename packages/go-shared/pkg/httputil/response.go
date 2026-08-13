package httputil

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/markettg/markettg/packages/go-shared/pkg/apperrors"
	"github.com/markettg/markettg/packages/go-shared/pkg/middleware"
)

func JSON(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(data)
}

func Error(c *fiber.Ctx, err error) error {
	requestID := middleware.GetRequestID(c)

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.StatusCode).JSON(apperrors.ErrorResponse{
			Error: *apperrors.Wrap(appErr, requestID),
		})
	}

	internal := apperrors.Wrap(apperrors.ErrInternal, requestID)
	return c.Status(internal.StatusCode).JSON(apperrors.ErrorResponse{Error: *internal})
}

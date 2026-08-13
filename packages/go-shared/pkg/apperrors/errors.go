package apperrors

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	RequestID  string `json:"request_id,omitempty"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: status}
}

var (
	ErrUnauthorized     = New("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	ErrForbidden        = New("FORBIDDEN", "Forbidden", http.StatusForbidden)
	ErrNotFound         = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrBadRequest       = New("BAD_REQUEST", "Bad request", http.StatusBadRequest)
	ErrConflict         = New("CONFLICT", "Conflict", http.StatusConflict)
	ErrInternal         = New("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrRateLimited      = New("RATE_LIMITED", "Too many requests", http.StatusTooManyRequests)
	ErrProductNotFound  = New("PRODUCT_NOT_FOUND", "Product not found", http.StatusNotFound)
	ErrOrderNotFound    = New("ORDER_NOT_FOUND", "Order not found", http.StatusNotFound)
	ErrPaymentNotFound  = New("PAYMENT_NOT_FOUND", "Payment not found", http.StatusNotFound)
	ErrInvalidPromoCode = New("INVALID_PROMO_CODE", "Invalid promo code", http.StatusBadRequest)
)

type ErrorResponse struct {
	Error AppError `json:"error"`
}

func Wrap(err *AppError, requestID string) *AppError {
	return &AppError{
		Code:       err.Code,
		Message:    err.Message,
		RequestID:  requestID,
		StatusCode: err.StatusCode,
	}
}

package errors

import (
	stdErrors "errors"
	"fmt"
	"net/http"
)

// As mengecek dan meng-unwrap error ke target pointer (wrapper dari errors.As bawaan Go)
func As(err error, target any) bool {
	return stdErrors.As(err, target)
}

// Is memeriksa apakah err cocok dengan target error (wrapper dari errors.Is bawaan Go)
func Is(err, target error) bool {
	return stdErrors.Is(err, target)
}

// Daftar kode error domain standar sesuai dokumentasi api.md
const (
	ErrValidation                  = "VALIDATION_ERROR"
	ErrUnauthorized                = "UNAUTHORIZED"
	ErrForbidden                   = "FORBIDDEN"
	ErrNotFound                    = "RESOURCE_NOT_FOUND"
	ErrInternal                    = "INTERNAL_SERVER_ERROR"
	ErrConflict                    = "CONFLICT"

	// RBAC
	ErrRoleNotFound                = "ROLE_NOT_FOUND"
	ErrPermissionNotFound          = "PERMISSION_NOT_FOUND"
	ErrRoleCannotBeDeleted         = "ROLE_CANNOT_BE_DELETED"

	// Trip & Items
	ErrTripNotOpen                 = "TRIP_NOT_OPEN"
	ErrOrderDeadlinePassed         = "ORDER_DEADLINE_PASSED"
	ErrItemNotAvailable            = "ITEM_NOT_AVAILABLE"
	ErrInsufficientQuota           = "INSUFFICIENT_QUOTA"

	// Orders & Fulfilment
	ErrInvalidOrderStatusTransition = "INVALID_ORDER_STATUS_TRANSITION"
	ErrOrderNotCancellable         = "ORDER_NOT_CANCELLABLE"

	// Payment
	ErrPaymentAlreadyConfirmed     = "PAYMENT_ALREADY_CONFIRMED"
	ErrIdempotencyKeyConflict      = "IDEMPOTENCY_KEY_CONFLICT"
)

// AppError adalah error kustom untuk kebutuhan business logic di layer service
type AppError struct {
	HTTPCode int                 `json:"http_code"`
	Code     string              `json:"code,omitempty"`
	Message  string              `json:"message"`
	Errors   map[string][]string `json:"errors,omitempty"`
}

func (e *AppError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	return e.Message
}

// New membuat instance AppError kustom baru
func New(httpCode int, code, message string) *AppError {
	return &AppError{
		HTTPCode: httpCode,
		Code:     code,
		Message:  message,
	}
}

// NewWithErrors membuat instance AppError dengan rincian error per-field (array of string)
func NewWithErrors(httpCode int, message string, errorsMap map[string][]string) *AppError {
	return &AppError{
		HTTPCode: httpCode,
		Code:     ErrValidation,
		Message:  message,
		Errors:   errorsMap,
	}
}

// Helper factory functions
func NewBadRequest(message string, errorsMap ...map[string][]string) *AppError {
	var errs map[string][]string
	if len(errorsMap) > 0 {
		errs = errorsMap[0]
	}
	return &AppError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrValidation,
		Message:  message,
		Errors:   errs,
	}
}

func NewUnauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, ErrUnauthorized, message)
}

func NewForbidden(message string) *AppError {
	return New(http.StatusForbidden, ErrForbidden, message)
}

func NewNotFound(message string) *AppError {
	return New(http.StatusNotFound, ErrNotFound, message)
}

func NewConflict(code, message string) *AppError {
	return New(http.StatusConflict, code, message)
}

func NewInternal(message string) *AppError {
	return New(http.StatusInternalServerError, ErrInternal, message)
}

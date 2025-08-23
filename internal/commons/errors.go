package commons

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenFailed        = errors.New("token generation failed")
	ErrUserNotActive      = errors.New("user is not active")
	ErrUserAlreadyActive  = errors.New("user is already active")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTokenExpired       = errors.New("token has expired")
)

var (
	ErrLinkNotFound     = errors.New("short link not found")
	ErrUnauthorized     = errors.New("unauthorized to access this link")
	ErrShortCodeExists  = errors.New("short code already exists")
	ErrInvalidShortCode = errors.New("invalid short code format")
)

var (
	ErrLinkInactive        = errors.New("link is inactive")
	ErrLinkExpired         = errors.New("link has expired")
	ErrClickLimitExceeded  = errors.New("click limit exceeded")
	ErrLinkAlreadyExists   = errors.New("link with this short code already exists")
	ErrLinkUpdateFailed    = errors.New("failed to update link")
	ErrLinkDeleteFailed    = errors.New("failed to delete link")
	ErrLinkCreateFailed    = errors.New("failed to create link")
	ErrLinkSearchFailed    = errors.New("failed to search links")
	ErrLinkClickFailed     = errors.New("failed to increment link clicks")
	ErrLinkDecrementFailed = errors.New("failed to decrement link click limit")
	ErrLinkNotActive       = errors.New("link is not active")
	ErrLinkNotOwnedByUser  = errors.New("link does not belong to the user")
)

// FieldError is a custom struct to hold detailed validation error information.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// FormatValidationErrors takes a validation error and returns a slice of FieldError.
func FormatValidationErrors(err error) []FieldError {
	var validationErrors validator.ValidationErrors

	// Check if the error is a validation error.
	if errors.As(err, &validationErrors) {
		// Create a slice to hold our custom error structs.
		out := make([]FieldError, len(validationErrors))
		for i, ve := range validationErrors {
			out[i] = FieldError{
				Field: ve.Field(),
				// You can customize the error message here.
				Error: fmt.Sprintf("This field failed the '%s' validation", ve.Tag()),
			}
		}
		return out
	}
	// If it's not a validation error, return nil.
	return nil
}

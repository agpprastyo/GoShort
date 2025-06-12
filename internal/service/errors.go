package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenFailed        = errors.New("token generation failed")
)

var (
	ErrLinkNotFound     = errors.New("short link not found")
	ErrUnauthorized     = errors.New("unauthorized to modify this link")
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

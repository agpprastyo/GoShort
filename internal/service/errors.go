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

package token

import (
	"GoShort/config"
	"GoShort/internal/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Common token-related errors
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims defines the custom claims for JWT
type Claims struct {
	UserID   string              `json:"user_id"`
	Username string              `json:"username"`
	Email    string              `json:"email"`
	Role     repository.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	config *config.AppConfig
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(config *config.AppConfig) *JWTMaker {
	return &JWTMaker{
		config: config,
	}
}

// GenerateToken creates a new token for a user
func (maker *JWTMaker) GenerateToken(user repository.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(maker.config.JWT.Expire)

	claims := Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    maker.config.JWT.Issuer,
			Audience:  []string{maker.config.JWT.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(maker.config.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// VerifyToken checks if the token is valid
func (maker *JWTMaker) VerifyToken(tokenString string) (*Claims, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.config.JWT.Secret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateTokenFromUserID creates a token using just user ID
func (maker *JWTMaker) GenerateTokenFromUserID(userID pgtype.UUID, role repository.UserRole) (string, time.Time, error) {
	expiresAt := time.Now().Add(maker.config.JWT.Expire)

	claims := Claims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    maker.config.JWT.Issuer,
			Audience:  []string{maker.config.JWT.Audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(maker.config.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

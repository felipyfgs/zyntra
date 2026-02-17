package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims
type Claims struct {
	UserID string    `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// DefaultJWTConfig returns config from environment
func DefaultJWTConfig() *JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "zyntra-dev-secret-change-in-production"
	}

	return &JWTConfig{
		SecretKey:          secret,
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour, // 7 days
		Issuer:             "zyntra",
	}
}

// JWTService handles JWT token operations
type JWTService struct {
	config *JWTConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(config *JWTConfig) *JWTService {
	if config == nil {
		config = DefaultJWTConfig()
	}
	return &JWTService{config: config}
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// GenerateTokenPair creates both access and refresh tokens
func (s *JWTService) GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
	accessToken, accessExp, err := s.generateToken(userID, email, role, AccessToken, s.config.AccessTokenExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.generateToken(userID, email, role, RefreshToken, s.config.RefreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		TokenType:    "Bearer",
	}, nil
}

// GenerateAccessToken creates only an access token
func (s *JWTService) GenerateAccessToken(userID, email, role string) (string, time.Time, error) {
	return s.generateToken(userID, email, role, AccessToken, s.config.AccessTokenExpiry)
}

// generateToken creates a JWT token
func (s *JWTService) generateToken(userID, email, role string, tokenType TokenType, expiry time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(expiry)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.Issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates and parses a JWT token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.SecretKey), nil
	})

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

// ValidateAccessToken validates an access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != AccessToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != RefreshToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshAccessToken creates a new access token from a valid refresh token
func (s *JWTService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return s.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
}

// internal/infrastructure/security/jwt_service.go
package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// JWTServiceImpl implementa JWTService
type JWTServiceImpl struct {
	secret               string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	issuer               string
	logger               *zap.Logger
}

// NewJWTService crea una nueva instancia de JWTServiceImpl
func NewJWTService(
	secret string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	issuer string,
	logger *zap.Logger,
) JWTService {
	return &JWTServiceImpl{
		secret:               secret,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		issuer:               issuer,
		logger:               logger,
	}
}

// AccessTokenClaims representa los claims MINIMALISTAS del access token
type AccessTokenClaims struct {
	UserID string `json:"sub"`  // Subject = UserID (estándar JWT)
	Role   string `json:"role"` // Solo role, necesario para autorización
	jwt.RegisteredClaims
}

// RefreshTokenClaims representa los claims del refresh token
type RefreshTokenClaims struct {
	UserID string `json:"sub"` // Solo UserID
	jwt.RegisteredClaims
}

// GenerateAccessToken genera un nuevo access token MINIMALISTA
func (s *JWTServiceImpl) GenerateAccessToken(
	userID, email, fullName, role string,
) (string, time.Time, error) {

	s.logger.Debug("Generando access token",
		zap.String("user_id", userID),
	)

	now := time.Now()
	expiresAt := now.Add(s.accessTokenDuration)

	claims := AccessTokenClaims{
		UserID: userID,
		Role:   role, // Solo guardamos role, es necesario para permisos
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error signing access token: %w", err)
	}

	s.logger.Info("Access token generado",
		zap.String("user_id", userID),
		zap.Time("expires_at", expiresAt),
	)

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken genera un nuevo refresh token MINIMALISTA
func (s *JWTServiceImpl) GenerateRefreshToken(userID string) (string, time.Time, string, error) {
	s.logger.Debug("Generando refresh token",
		zap.String("user_id", userID),
	)

	now := time.Now()
	expiresAt := now.Add(s.refreshTokenDuration)
	tokenID := uuid.New().String()

	claims := RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("error signing refresh token: %w", err)
	}

	s.logger.Info("Refresh token generado",
		zap.String("user_id", userID),
		zap.Time("expires_at", expiresAt),
	)

	return tokenString, expiresAt, tokenID, nil
}

// ValidateAccessToken valida un access token y retorna userID, role y jti
func (s *JWTServiceImpl) ValidateAccessToken(tokenString string) (string, string, string, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return "", "", "", fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return "", "", "", errors.New("invalid token")
	}

	return claims.UserID, claims.Role, claims.ID, nil
}

// GetAccessTokenExpiration parsea un access token y retorna su fecha de expiración
func (s *JWTServiceImpl) GetAccessTokenExpiration(tokenString string) (time.Time, error) {
	claims := &AccessTokenClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing token expiration: %w", err)
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("token sin fecha de expiración")
	}

	return claims.ExpiresAt.Time, nil
}

// ValidateRefreshToken valida un refresh token
func (s *JWTServiceImpl) ValidateRefreshToken(tokenString string) (string, string, error) {
	claims := &RefreshTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return "", "", fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return "", "", errors.New("invalid token")
	}

	return claims.UserID, claims.ID, nil
}

// Asegurar que implementa la interfaz
var _ JWTService = (*JWTServiceImpl)(nil)

package tokenmanager

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	config TokenManagerConfig
}

func NewTokenManager(config TokenManagerConfig) *TokenManager {
	return &TokenManager{
		config: config,
	}
}

// GenerateAccessToken генерируем access-token
func (tm *TokenManager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(tm.config.accessTokenTTL).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.config.jwtSecret)
}

func (tm *TokenManager) GetRefreshTokenTTL() time.Duration {
	return tm.config.refreshTokenTTL
}

func (tm *TokenManager) GetAccessTokenTTL() time.Duration {
	return tm.config.accessTokenTTL
}

// GenerateRefreshToken генерируем refresh-token
func (tm *TokenManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 24)
	_, err := rand.Read(b) // заполняем токен
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil // приводим в вид ascii строки
	//  для безопасной передачи
}

// HashToken хеширует рефреш-токен (SHA-256 + hex)
func (tm *TokenManager) HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ErrValidToken ошибка при валидации токена
var ErrInvalidToken = errors.New("invalid token")

// ParseToken парсит access-токен и возвращаеи uuid пользователя при успешной валидации
// TO-DO: если access-token насытится большими данными,
// возвращать кастомную структуру claims
func (tm *TokenManager) ParseToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return tm.config.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// ValidateToken просто проверяет валидность токена
func (tm *TokenManager) ValidateToken(tokenString string) error {
	_, err := tm.ParseToken(tokenString)
	return err
}

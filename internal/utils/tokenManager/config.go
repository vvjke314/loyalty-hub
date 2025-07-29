package tokenmanager

import (
	"os"
	"time"
)

const (
	defaultAccessTokenTTL  time.Duration = 20 * time.Minute
	defaultRefreshTokenTTL time.Duration = 30 * 24 * time.Hour
)

type TokenManagerConfigOption interface {
	apply(*TokenManagerConfig)
}

type AccessTokenTTLOption struct {
	accessTokenTTL time.Duration
}

func WithAccessTokenTTL(ttl time.Duration) TokenManagerConfigOption {
	return AccessTokenTTLOption{
		accessTokenTTL: ttl,
	}
}

func (o AccessTokenTTLOption) apply(cfg *TokenManagerConfig) {
	cfg.accessTokenTTL = o.accessTokenTTL
}

type RefreshTokenTTLOption struct {
	refreshTokenTTL time.Duration
}

func WithRefreshTokenTTL(ttl time.Duration) TokenManagerConfigOption {
	return RefreshTokenTTLOption{
		refreshTokenTTL: ttl,
	}
}

func (o RefreshTokenTTLOption) apply(cfg *TokenManagerConfig) {
	cfg.refreshTokenTTL = o.refreshTokenTTL
}

type TokenManagerConfig struct {
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenManagerConfig(opts ...TokenManagerConfigOption) TokenManagerConfig {
	cfg := &TokenManagerConfig{
		jwtSecret:       []byte(os.Getenv("JWT_SECRET")),
		accessTokenTTL:  defaultAccessTokenTTL,
		refreshTokenTTL: defaultRefreshTokenTTL,
	}

	for _, o := range opts {
		o.apply(cfg)
	}

	return *cfg
}

package config

import "context"

type contextKey int

var configContextKey contextKey = 0

func FromContext(ctx context.Context) *Config {
	cfg, _ := ctx.Value(configContextKey).(*Config)
	return cfg
}

func NewContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configContextKey, cfg)
}

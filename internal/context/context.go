package context

import (
	"context"

	"github.com/CloudNativeWorks/elchi-discovery/internal/config"
)

type contextKey string

const configKey contextKey = "elchi-config"

func WithConfig(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

func GetConfig(ctx context.Context) *config.Config {
	if cfg, ok := ctx.Value(configKey).(*config.Config); ok {
		return cfg
	}
	return nil
}

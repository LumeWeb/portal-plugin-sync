package service

import (
	"go.lumeweb.com/portal/config"
)

var _ config.ServiceConfig = (*ServiceConfig)(nil)

type ServiceConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func (s ServiceConfig) Defaults() map[string]any {
	return map[string]any{
		"enabled": false,
	}
}

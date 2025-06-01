package grpc

import (
	"fmt"
	"time"

	"distributed-service/pkg/config"
)

// ConvertConfig converts config.GRPCConfig to grpc.Config
func ConvertConfig(cfg *config.GRPCConfig) (*Config, error) {
	connectionTimeout, err := time.ParseDuration(cfg.ConnectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid connection_timeout: %w", err)
	}

	maxConnectionIdle, err := time.ParseDuration(cfg.MaxConnectionIdle)
	if err != nil {
		return nil, fmt.Errorf("invalid max_connection_idle: %w", err)
	}

	maxConnectionAge, err := time.ParseDuration(cfg.MaxConnectionAge)
	if err != nil {
		return nil, fmt.Errorf("invalid max_connection_age: %w", err)
	}

	maxConnectionAgeGrace, err := time.ParseDuration(cfg.MaxConnectionAgeGrace)
	if err != nil {
		return nil, fmt.Errorf("invalid max_connection_age_grace: %w", err)
	}

	timeVal, err := time.ParseDuration(cfg.Time)
	if err != nil {
		return nil, fmt.Errorf("invalid time: %w", err)
	}

	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %w", err)
	}

	return &Config{
		Port:                  cfg.Port,
		MaxRecvMsgSize:        cfg.MaxRecvMsgSize,
		MaxSendMsgSize:        cfg.MaxSendMsgSize,
		ConnectionTimeout:     connectionTimeout,
		MaxConnectionIdle:     maxConnectionIdle,
		MaxConnectionAge:      maxConnectionAge,
		MaxConnectionAgeGrace: maxConnectionAgeGrace,
		Time:                  timeVal,
		Timeout:               timeout,
		EnableReflection:      cfg.EnableReflection,
		EnableHealthCheck:     cfg.EnableHealthCheck,
	}, nil
}

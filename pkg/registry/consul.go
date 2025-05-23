package registry

import (
	"context"
	"distributed-service/pkg/config"
	"distributed-service/pkg/logger"
	"fmt"

	"github.com/hashicorp/consul/api"
)

type ServiceRegistry struct {
	client *api.Client
	cfg    *config.ConsulConfig
}

func NewServiceRegistry(ctx context.Context, cfg *config.ConsulConfig) (*ServiceRegistry, error) {
	logger.Info(ctx, "Initializing Consul client",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
	)

	consulConfig := api.DefaultConfig()
	consulConfig.Address = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client, err := api.NewClient(consulConfig)
	if err != nil {
		logger.Error(ctx, "Failed to create Consul client",
			logger.Error_(err),
			logger.String("address", consulConfig.Address),
		)
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}

	// Test connection
	_, err = client.Status().Leader()
	if err != nil {
		logger.Error(ctx, "Failed to connect to Consul",
			logger.Error_(err),
			logger.String("address", consulConfig.Address),
		)
		return nil, fmt.Errorf("failed to connect to Consul: %w", err)
	}

	logger.Info(ctx, "Successfully connected to Consul",
		logger.String("address", consulConfig.Address),
	)

	return &ServiceRegistry{
		client: client,
		cfg:    cfg,
	}, nil
}

func (sr *ServiceRegistry) RegisterService(ctx context.Context, serverCfg *config.ServerConfig) error {
	logger.Info(ctx, "Registering service with Consul",
		logger.String("service", serverCfg.Name),
		logger.String("version", serverCfg.Version),
		logger.Int("port", serverCfg.Port),
	)

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%s-%s-%d", serverCfg.Name, serverCfg.Version, serverCfg.Port),
		Name:    serverCfg.Name,
		Port:    serverCfg.Port,
		Tags:    []string{serverCfg.Version},
		Address: "localhost",
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://localhost:%d/health", serverCfg.Port),
			Interval:                       sr.cfg.ServiceCheckInterval,
			DeregisterCriticalServiceAfter: sr.cfg.DeregisterCriticalServiceAfter,
		},
	}

	if err := sr.client.Agent().ServiceRegister(registration); err != nil {
		logger.Error(ctx, "Failed to register service",
			logger.Error_(err),
			logger.String("service", serverCfg.Name),
		)
		return fmt.Errorf("failed to register service: %w", err)
	}

	logger.Info(ctx, "Successfully registered service",
		logger.String("service", serverCfg.Name),
		logger.String("id", registration.ID),
	)

	return nil
}

func (sr *ServiceRegistry) DeregisterService(ctx context.Context, serverCfg *config.ServerConfig) error {
	serviceID := fmt.Sprintf("%s-%s-%d", serverCfg.Name, serverCfg.Version, serverCfg.Port)

	logger.Info(ctx, "Deregistering service from Consul",
		logger.String("service", serverCfg.Name),
		logger.String("id", serviceID),
	)

	if err := sr.client.Agent().ServiceDeregister(serviceID); err != nil {
		logger.Error(ctx, "Failed to deregister service",
			logger.Error_(err),
			logger.String("service", serverCfg.Name),
			logger.String("id", serviceID),
		)
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	logger.Info(ctx, "Successfully deregistered service",
		logger.String("service", serverCfg.Name),
		logger.String("id", serviceID),
	)

	return nil
}

func (sr *ServiceRegistry) GetService(ctx context.Context, name string) ([]*api.ServiceEntry, error) {
	logger.Info(ctx, "Looking up service",
		logger.String("service", name),
	)

	services, _, err := sr.client.Health().Service(name, "", true, nil)
	if err != nil {
		logger.Error(ctx, "Failed to get service",
			logger.Error_(err),
			logger.String("service", name),
		)
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	logger.Info(ctx, "Found service instances",
		logger.String("service", name),
		logger.Int("count", len(services)),
	)

	return services, nil
}

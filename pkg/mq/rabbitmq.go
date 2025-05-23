package mq

import (
	"context"
	"distributed-service/pkg/config"
	"distributed-service/pkg/logger"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitMQConn *amqp.Connection
var RabbitMQChannel *amqp.Channel

func InitRabbitMQ(ctx context.Context, cfg *config.RabbitMQConfig) error {
	logger.Info(ctx, "Initializing RabbitMQ connection",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("vhost", cfg.VHost),
	)

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.VHost,
	)

	// Set connection timeout
	dialConfig := amqp.Config{
		Dial: amqp.DefaultDial(5 * time.Second),
	}

	conn, err := amqp.DialConfig(url, dialConfig)
	if err != nil {
		logger.Error(ctx, "Failed to connect to RabbitMQ",
			logger.Error_(err),
			logger.String("host", cfg.Host),
			logger.Int("port", cfg.Port),
		)
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Error(ctx, "Failed to open RabbitMQ channel", logger.Error_(err))
		return fmt.Errorf("failed to open channel: %w", err)
	}

	logger.Info(ctx, "Successfully connected to RabbitMQ",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.String("vhost", cfg.VHost),
	)

	RabbitMQConn = conn
	RabbitMQChannel = ch

	// Setup connection monitoring
	go monitorConnection(ctx, cfg)

	return nil
}

func CloseRabbitMQ(ctx context.Context) {
	if RabbitMQChannel != nil {
		if err := RabbitMQChannel.Close(); err != nil {
			logger.Error(ctx, "Error closing RabbitMQ channel", logger.Error_(err))
		}
	}
	if RabbitMQConn != nil {
		if err := RabbitMQConn.Close(); err != nil {
			logger.Error(ctx, "Error closing RabbitMQ connection", logger.Error_(err))
		}
	}
	logger.Info(ctx, "RabbitMQ connection closed")
}

func monitorConnection(ctx context.Context, cfg *config.RabbitMQConfig) {
	for {
		reason, ok := <-RabbitMQConn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			logger.Info(ctx, "RabbitMQ connection closed")
			return
		}

		logger.Error(ctx, "RabbitMQ connection lost",
			logger.String("reason", reason.Error()),
		)

		// Attempt to reconnect
		for {
			time.Sleep(5 * time.Second)

			if err := InitRabbitMQ(ctx, cfg); err != nil {
				logger.Error(ctx, "Failed to reconnect to RabbitMQ", logger.Error_(err))
				continue
			}

			logger.Info(ctx, "Successfully reconnected to RabbitMQ")
			break
		}
	}
}

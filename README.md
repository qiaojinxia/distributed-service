# Distributed Microservices Framework

A Go-based distributed microservices framework with integrated MySQL, Redis, and RabbitMQ support.

## Features

- Configuration management using Viper
- MySQL integration with GORM
- Redis integration
- RabbitMQ integration
- Health check endpoints
- Easy to extend and customize

## Prerequisites

- Go 1.21 or later
- MySQL
- Redis
- RabbitMQ

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd distributed-service
```

2. Install dependencies:
```bash
go mod tidy
```

3. Configure the services:
Edit `config/config.yaml` to match your environment settings.

4. Run the application:
```bash
go run main.go
```

## Configuration

The application uses a YAML configuration file located at `config/config.yaml`. You can customize the following settings:

- Server configuration (port, mode)
- MySQL connection details
- Redis connection details
- RabbitMQ connection details

## API Endpoints

### Health Check
```
GET /health
```
Returns the status of all connected services.

## Project Structure

```
.
├─ config/
│   └── config.yaml
├── pkg/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── mysql.go
│   │   └── redis.go
│   └── mq/
│       └── rabbitmq.go
├── go.mod
├── main.go
└── README.md
```

## Contributing

Feel free to submit issues and enhancement requests! 
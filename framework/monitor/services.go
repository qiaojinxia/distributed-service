package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"distributed-service/framework/logger"

	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/consul/api"
	amqp091 "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`    // "healthy", "unhealthy", "unknown"
	Message   string                 `json:"message"`   // Status message
	Latency   int64                  `json:"latency"`   // Response latency in milliseconds
	Timestamp time.Time              `json:"timestamp"` // Last check timestamp
	Details   map[string]interface{} `json:"details"`   // Detailed service-specific information
}

// ServicesStats represents overall service statistics
type ServicesStats struct {
	Services []ServiceStatus `json:"services"`
	Summary  ServiceSummary  `json:"summary"`
}

// ServiceSummary provides a summary of service health
type ServiceSummary struct {
	Total     int `json:"total"`     // Total number of services
	Healthy   int `json:"healthy"`   // Number of healthy services
	Unhealthy int `json:"unhealthy"` // Number of unhealthy services
	Unknown   int `json:"unknown"`   // Number of unknown status services
}

// ServiceConfig holds configuration for service monitoring
type ServiceConfig struct {
	MySQL    MySQLConfig    `json:"mysql"`
	Redis    RedisConfig    `json:"redis"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`
	Consul   ConsulConfig   `json:"consul"`
	GRPC     GRPCConfig     `json:"grpc"`
}

type MySQLConfig struct {
	DSN     string `json:"dsn"`
	Enabled bool   `json:"enabled"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	Enabled  bool   `json:"enabled"`
}

type RabbitMQConfig struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type ConsulConfig struct {
	Address string `json:"address"`
	Enabled bool   `json:"enabled"`
}

type GRPCConfig struct {
	Address string `json:"address"`
	Enabled bool   `json:"enabled"`
}

// ServiceMonitor provides service monitoring capabilities
type ServiceMonitor struct {
	ctx    context.Context
	config ServiceConfig
}

// NewServiceMonitor creates a new service monitor
func NewServiceMonitor(ctx context.Context, config ServiceConfig) *ServiceMonitor {
	return &ServiceMonitor{
		ctx:    ctx,
		config: config,
	}
}

// GetServicesStats retrieves status for all configured services
func (sm *ServiceMonitor) GetServicesStats() (*ServicesStats, error) {
	var services []ServiceStatus

	// Check MySQL
	if sm.config.MySQL.Enabled {
		status := sm.checkMySQL()
		services = append(services, status)
	}

	// Check Redis
	if sm.config.Redis.Enabled {
		status := sm.checkRedis()
		services = append(services, status)
	}

	// Check RabbitMQ
	if sm.config.RabbitMQ.Enabled {
		status := sm.checkRabbitMQ()
		services = append(services, status)
	}

	// Check Consul
	if sm.config.Consul.Enabled {
		status := sm.checkConsul()
		services = append(services, status)
	}

	// Check gRPC
	if sm.config.GRPC.Enabled {
		status := sm.checkGRPC()
		services = append(services, status)
	}

	// Calculate summary
	summary := sm.calculateSummary(services)

	return &ServicesStats{
		Services: services,
		Summary:  summary,
	}, nil
}

// checkMySQL checks MySQL database connectivity
func (sm *ServiceMonitor) checkMySQL() ServiceStatus {
	start := time.Now()
	status := ServiceStatus{
		Name:      "MySQL",
		Status:    "unknown",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if sm.config.MySQL.DSN == "" {
		status.Status = "unhealthy"
		status.Message = "DSN not configured"
		return status
	}

	db, err := sql.Open("mysql", sm.config.MySQL.DSN)
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Failed to open connection: %v", err)
		return status
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	// Set connection limits for the check
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Second)

	ctx, cancel := context.WithTimeout(sm.ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Ping failed: %v", err)
		return status
	}

	// Collect detailed connection pool statistics
	stats := db.Stats()
	status.Details["connection_pool"] = map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration_ns":     stats.WaitDuration.Nanoseconds(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}

	// Test a simple query for additional validation
	var version string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		status.Status = "degraded"
		status.Message = fmt.Sprintf("Query test failed: %v", err)
		status.Details["query_test"] = "failed"
	} else {
		status.Status = "healthy"
		status.Message = "Connection successful"
		status.Details["mysql_version"] = version
		status.Details["query_test"] = "passed"
	}

	status.Latency = time.Since(start).Milliseconds()
	status.Details["dsn_info"] = map[string]interface{}{
		"masked_dsn": maskPassword(sm.config.MySQL.DSN),
	}

	logger.Info(sm.ctx, "MySQL health check completed",
		zap.String("status", status.Status),
		zap.Int64("latency_ms", status.Latency))

	return status
}

// checkRedis checks Redis connectivity
func (sm *ServiceMonitor) checkRedis() ServiceStatus {
	start := time.Now()
	status := ServiceStatus{
		Name:      "Redis",
		Status:    "unknown",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if sm.config.Redis.Addr == "" {
		status.Status = "unhealthy"
		status.Message = "Redis address not configured"
		return status
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     sm.config.Redis.Addr,
		Password: sm.config.Redis.Password,
		DB:       sm.config.Redis.DB,
	})
	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			_ = rdb.Close()
		}
	}(rdb)

	ctx, cancel := context.WithTimeout(sm.ctx, 5*time.Second)
	defer cancel()

	// Basic ping test
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Ping failed: %v", err)
		return status
	}

	// Collect detailed Redis information
	status.Details["ping_response"] = pong

	// Get Redis info
	info, err := rdb.Info(ctx).Result()
	if err != nil {
		status.Status = "degraded"
		status.Message = "Connection successful but info collection failed"
		status.Details["info_error"] = err.Error()
	} else {
		status.Status = "healthy"
		status.Message = "Connection successful"

		// Parse Redis info for key metrics
		redisInfo := parseRedisInfo(info)
		status.Details["redis_info"] = redisInfo
	}

	// Get connection pool stats
	poolStats := rdb.PoolStats()
	status.Details["connection_pool"] = map[string]interface{}{
		"hits":        poolStats.Hits,
		"misses":      poolStats.Misses,
		"timeouts":    poolStats.Timeouts,
		"total_conns": poolStats.TotalConns,
		"idle_conns":  poolStats.IdleConns,
		"stale_conns": poolStats.StaleConns,
	}

	// Test basic operations
	testKey := fmt.Sprintf("health_check_%d", time.Now().Unix())
	if err := rdb.Set(ctx, testKey, "test", time.Second*10).Err(); err != nil {
		status.Details["write_test"] = "failed"
		status.Details["write_error"] = err.Error()
	} else {
		status.Details["write_test"] = "passed"

		// Test read
		if val, err := rdb.Get(ctx, testKey).Result(); err != nil {
			status.Details["read_test"] = "failed"
			status.Details["read_error"] = err.Error()
		} else {
			status.Details["read_test"] = "passed"
			status.Details["read_value"] = val
		}

		// Cleanup test key
		rdb.Del(ctx, testKey)
	}

	status.Latency = time.Since(start).Milliseconds()
	status.Details["connection_info"] = map[string]interface{}{
		"address":  sm.config.Redis.Addr,
		"database": sm.config.Redis.DB,
	}

	logger.Info(sm.ctx, "Redis health check completed",
		zap.String("status", status.Status),
		zap.Int64("latency_ms", status.Latency))

	return status
}

// checkRabbitMQ checks RabbitMQ connectivity
func (sm *ServiceMonitor) checkRabbitMQ() ServiceStatus {
	start := time.Now()
	status := ServiceStatus{
		Name:      "RabbitMQ",
		Status:    "unknown",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if sm.config.RabbitMQ.URL == "" {
		status.Status = "unhealthy"
		status.Message = "RabbitMQ URL not configured"
		return status
	}

	conn, err := amqp091.Dial(sm.config.RabbitMQ.URL)
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Connection failed: %v", err)
		status.Details["connection_error"] = err.Error()
		return status
	}
	defer func(conn *amqp091.Connection) {
		_ = conn.Close()
	}(conn)

	// Test channel creation
	ch, err := conn.Channel()
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Channel creation failed: %v", err)
		status.Details["channel_error"] = err.Error()
		return status
	}
	defer func(ch *amqp091.Channel) {
		_ = ch.Close()
	}(ch)

	// Collect connection details
	status.Details["connection_info"] = map[string]interface{}{
		"url_masked":       maskRabbitMQURL(sm.config.RabbitMQ.URL),
		"local_addr":       conn.LocalAddr().String(),
		"remote_addr":      conn.RemoteAddr().String(),
		"connection_state": conn.IsClosed(),
	}

	// Test basic queue operations
	testQueue := fmt.Sprintf("health_check_%d", time.Now().Unix())

	// Declare a temporary queue
	q, err := ch.QueueDeclare(
		testQueue, // name
		false,     // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		status.Status = "degraded"
		status.Message = "Connection successful but queue operations failed"
		status.Details["queue_test"] = "failed"
		status.Details["queue_error"] = err.Error()
	} else {
		status.Details["queue_test"] = "passed"
		status.Details["test_queue"] = map[string]interface{}{
			"name":      q.Name,
			"messages":  q.Messages,
			"consumers": q.Consumers,
		}

		// Test publish
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp091.Publishing{
				ContentType: "text/plain",
				Body:        []byte("health check message"),
			})

		if err != nil {
			status.Details["publish_test"] = "failed"
			status.Details["publish_error"] = err.Error()
		} else {
			status.Details["publish_test"] = "passed"
		}

		// Clean up test queue
		_, _ = ch.QueueDelete(testQueue, false, false, false)
	}

	status.Status = "healthy"
	status.Message = "Connection successful"
	status.Latency = time.Since(start).Milliseconds()

	logger.Info(sm.ctx, "RabbitMQ health check completed",
		zap.String("status", status.Status),
		zap.Int64("latency_ms", status.Latency))

	return status
}

// checkConsul checks Consul connectivity
func (sm *ServiceMonitor) checkConsul() ServiceStatus {
	start := time.Now()
	status := ServiceStatus{
		Name:      "Consul",
		Status:    "unknown",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if sm.config.Consul.Address == "" {
		status.Status = "unhealthy"
		status.Message = "Consul address not configured"
		return status
	}

	config := api.DefaultConfig()
	config.Address = sm.config.Consul.Address

	client, err := api.NewClient(config)
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Client creation failed: %v", err)
		return status
	}

	// Try to get agent info
	_, err = client.Agent().Self()
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Agent info failed: %v", err)
		return status
	}

	status.Status = "healthy"
	status.Message = "Connection successful"
	status.Latency = time.Since(start).Milliseconds()
	status.Details["connection_string"] = sm.config.Consul.Address

	logger.Info(sm.ctx, "Consul health check completed",
		zap.String("status", status.Status),
		zap.Int64("latency_ms", status.Latency))

	return status
}

// checkGRPC checks gRPC server connectivity using health check protocol
func (sm *ServiceMonitor) checkGRPC() ServiceStatus {
	start := time.Now()
	status := ServiceStatus{
		Name:      "gRPC",
		Status:    "unknown",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	if sm.config.GRPC.Address == "" {
		status.Status = "unhealthy"
		status.Message = "gRPC address not configured"
		return status
	}

	// Create context with timeout for connection
	ctx, cancel := context.WithTimeout(sm.ctx, 5*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(
		sm.config.GRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("Connection failed: %v", err)
		status.Details["connection_error"] = err.Error()
		return status
	}
	// 2. 带上下文的连接（支持超时/取消）
	conn.Connect() // 替代 WithBlock 的阻塞行为

	defer func(conn *grpc.ClientConn) {
		err = conn.Close()
	}(conn)

	// Wait for connection to be ready
	if !conn.WaitForStateChange(ctx, connectivity.Connecting) {
		status.Status = "unhealthy"
		status.Message = "Connection timeout"
		status.Details["connection_state"] = conn.GetState().String()
		return status
	}

	// Collect connection details
	state := conn.GetState()
	status.Details["connection_state"] = state.String()
	status.Details["target"] = conn.Target()
	status.Details["connection_info"] = map[string]interface{}{
		"address": sm.config.GRPC.Address,
		"state":   state.String(),
	}

	// Check connection state
	if state != connectivity.Ready {
		// Try using health check protocol
		healthClient := grpc_health_v1.NewHealthClient(conn)

		// Create a new context for health check
		healthCtx, healthCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer healthCancel()

		resp, err := healthClient.Check(healthCtx, &grpc_health_v1.HealthCheckRequest{
			Service: "", // Empty service name for overall health
		})

		if err != nil {
			// If health check fails, fall back to basic connectivity check
			status.Status = "degraded"
			status.Message = fmt.Sprintf("Health check failed, connection state: %v", state)
			status.Details["health_check_error"] = err.Error()
			status.Details["health_check_available"] = false
			status.Latency = time.Since(start).Milliseconds()
			return status
		}

		status.Details["health_check_available"] = true
		status.Details["health_response"] = resp.Status.String()

		if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
			status.Status = "healthy"
			status.Message = "Health check passed"
		} else {
			status.Status = "unhealthy"
			status.Message = fmt.Sprintf("Service not serving: %v", resp.Status)
		}
	} else {
		// Connection is ready, try health check first
		healthClient := grpc_health_v1.NewHealthClient(conn)

		healthCtx, healthCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer healthCancel()

		resp, err := healthClient.Check(healthCtx, &grpc_health_v1.HealthCheckRequest{
			Service: "", // Empty service name for overall health
		})

		if err != nil {
			// Health check not available, but connection is ready
			status.Status = "healthy"
			status.Message = "Connection ready (health check not available)"
			status.Details["health_check_available"] = false
			status.Details["health_check_error"] = err.Error()
		} else {
			status.Details["health_check_available"] = true
			status.Details["health_response"] = resp.Status.String()

			if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
				status.Status = "healthy"
				status.Message = "Health check passed"
			} else {
				status.Status = "unhealthy"
				status.Message = fmt.Sprintf("Service not serving: %v", resp.Status)
			}
		}
	}

	status.Latency = time.Since(start).Milliseconds()

	logger.Info(sm.ctx, "gRPC health check completed",
		zap.String("status", status.Status),
		zap.Int64("latency_ms", status.Latency),
		zap.String("message", status.Message))

	return status
}

// calculateSummary calculates service health summary
func (sm *ServiceMonitor) calculateSummary(services []ServiceStatus) ServiceSummary {
	summary := ServiceSummary{
		Total: len(services),
	}

	for _, service := range services {
		switch service.Status {
		case "healthy":
			summary.Healthy++
		case "unhealthy":
			summary.Unhealthy++
		default:
			summary.Unknown++
		}
	}

	return summary
}

// CheckHTTPEndpoint checks if an HTTP endpoint is reachable
func (sm *ServiceMonitor) CheckHTTPEndpoint(name, url string, timeout time.Duration) ServiceStatus {
	start := time.Now()
	status := ServiceStatus{
		Name:      name,
		Status:    "unknown",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("HTTP request failed: %v", err)
		return status
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status.Status = "healthy"
		status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	status.Latency = time.Since(start).Milliseconds()
	return status
}

// CheckTCPEndpoint checks if a TCP endpoint is reachable
func (sm *ServiceMonitor) CheckTCPEndpoint(name, address string, timeout time.Duration) ServiceStatus {
	start := time.Now()
	status := ServiceStatus{
		Name:      name,
		Status:    "unknown",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("TCP connection failed: %v", err)
		return status
	}
	defer func(conn net.Conn) {
		err = conn.Close()
	}(conn)

	status.Status = "healthy"
	status.Message = "TCP connection successful"
	status.Latency = time.Since(start).Milliseconds()
	return status
}

// maskPassword masks sensitive information in connection strings
func maskPassword(connectionString string) string {
	// Simple password masking for MySQL DSN format: user:password@tcp(host:port)/database
	parts := strings.Split(connectionString, "@")
	if len(parts) < 2 {
		return connectionString
	}

	userPart := parts[0]
	if strings.Contains(userPart, ":") {
		userCredentials := strings.Split(userPart, ":")
		if len(userCredentials) >= 2 {
			userCredentials[1] = "****"
			parts[0] = strings.Join(userCredentials, ":")
		}
	}

	return strings.Join(parts, "@")
}

// parseRedisInfo parses Redis INFO output and extracts key metrics
func parseRedisInfo(info string) map[string]interface{} {
	result := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Extract important metrics
		switch key {
		case "redis_version", "redis_mode", "os", "arch_bits":
			result[key] = value
		case "uptime_in_seconds", "connected_clients", "used_memory", "used_memory_rss",
			"used_memory_peak", "total_connections_received", "total_commands_processed",
			"keyspace_hits", "keyspace_misses":
			if intVal, err := fmt.Sscanf(value, "%d", new(int64)); err == nil && intVal == 1 {
				var parsedValue int64
				_, _ = fmt.Sscanf(value, "%d", &parsedValue)
				result[key] = parsedValue
			} else {
				result[key] = value
			}
		}
	}

	return result
}

// maskRabbitMQURL masks sensitive information in RabbitMQ connection strings
func maskRabbitMQURL(url string) string {
	// Simple password masking for RabbitMQ connection strings
	parts := strings.Split(url, "@")
	if len(parts) < 2 {
		return url
	}

	userPart := parts[0]
	if strings.Contains(userPart, ":") {
		userCredentials := strings.Split(userPart, ":")
		if len(userCredentials) >= 2 {
			userCredentials[1] = "****"
			parts[0] = strings.Join(userCredentials, ":")
		}
	}

	return strings.Join(parts, "@")
}

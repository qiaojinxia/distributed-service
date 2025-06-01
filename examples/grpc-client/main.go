package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "distributed-service/api/proto/user"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func initTracing() func() {
	// Create a stdout exporter for demo purposes
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("Failed to create trace exporter: %v", err)
	}

	// Create a trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("grpc-client-example"),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
	)

	// Set the global trace provider
	otel.SetTracerProvider(tp)

	// Set the global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return a cleanup function
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}

// grpcTracingInterceptor creates a client-side tracing interceptor
func grpcTracingInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Start a span for the gRPC call
		tracer := otel.Tracer("grpc-client")
		ctx, span := tracer.Start(ctx, "gRPC Call: "+method)
		defer span.End()

		// Add attributes
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("component", "grpc-client"),
		)

		// Inject tracing context into metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Create a propagator to inject trace context
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, &metadataCarrier{md})

		// Add the metadata to the context
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Make the call
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}

// metadataCarrier implements TextMapCarrier for gRPC metadata
type metadataCarrier struct {
	md metadata.MD
}

func (mc *metadataCarrier) Get(key string) string {
	vals := mc.md.Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (mc *metadataCarrier) Set(key, value string) {
	mc.md.Set(key, value)
}

func (mc *metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.md))
	for key := range mc.md {
		keys = append(keys, key)
	}
	return keys
}

func main() {
	// Initialize tracing
	cleanup := initTracing()
	defer cleanup()

	// Create root span for the entire client session
	tracer := otel.Tracer("grpc-client")
	ctx, rootSpan := tracer.Start(context.Background(), "gRPC Client Example")
	defer rootSpan.End()

	// Connect to gRPC server with tracing interceptor
	conn, err := grpc.NewClient("localhost:9090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcTracingInterceptor()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Failed to close gRPC client connection: %v", err)
		}
	}(conn)

	// Create client
	client := pb.NewUserServiceClient(conn)
	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Test health check
	fmt.Println("=== Testing Health Check ===")
	_, span := tracer.Start(callCtx, "Health Check Test")
	healthResp, err := client.Check(callCtx, &pb.HealthCheckRequest{
		Service: "user.v1.UserService",
	})
	if err != nil {
		log.Printf("Health check failed: %v", err)
		span.RecordError(err)
	} else {
		fmt.Printf("Health check status: %v\n", healthResp.Status)
		span.SetAttributes(attribute.String("health.status", healthResp.Status.String()))
	}
	span.End()

	// Test user creation
	fmt.Println("\n=== Testing User Creation ===")
	_, span = tracer.Start(callCtx, "Create User Test")
	span.SetAttributes(
		attribute.String("user.username", "testuser"),
		attribute.String("user.email", "test@example.com"),
	)
	createResp, err := client.CreateUser(callCtx, &pb.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		log.Printf("Create user failed: %v", err)
		span.RecordError(err)
	} else {
		fmt.Printf("Created user: %+v\n", createResp.User)
		span.SetAttributes(attribute.Int64("user.created_id", int64(createResp.User.Id)))
	}
	span.End()

	// Test user login
	fmt.Println("\n=== Testing User Login ===")
	_, span = tracer.Start(callCtx, "Login Test")
	span.SetAttributes(attribute.String("user.username", "testuser"))
	loginResp, err := client.Login(callCtx, &pb.LoginRequest{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		log.Printf("Login failed: %v", err)
		span.RecordError(err)
	} else {
		fmt.Printf("Login successful!\n")
		fmt.Printf("Access Token: %s\n", loginResp.AccessToken[:20]+"...")
		fmt.Printf("User: %+v\n", loginResp.User)
		span.SetAttributes(
			attribute.Bool("login.success", true),
			attribute.Int64("user.id", int64(loginResp.User.Id)),
		)
	}
	span.End()

	// Test get user
	if createResp != nil && createResp.User != nil {
		fmt.Println("\n=== Testing Get User ===")
		_, span = tracer.Start(callCtx, "Get User Test")
		span.SetAttributes(attribute.Int64("user.id", int64(createResp.User.Id)))
		getUserResp, err := client.GetUser(callCtx, &pb.GetUserRequest{
			Id: createResp.User.Id,
		})
		if err != nil {
			log.Printf("Get user failed: %v", err)
			span.RecordError(err)
		} else {
			fmt.Printf("Retrieved user: %+v\n", getUserResp.User)
			span.SetAttributes(attribute.String("user.username", getUserResp.User.Username))
		}
		span.End()

		// Test update user
		fmt.Println("\n=== Testing Update User ===")
		_, span = tracer.Start(callCtx, "Update User Test")
		span.SetAttributes(
			attribute.Int64("user.id", int64(createResp.User.Id)),
			attribute.String("user.new_username", "updateduser"),
		)
		updateResp, err := client.UpdateUser(callCtx, &pb.UpdateUserRequest{
			Id:       createResp.User.Id,
			Username: "updateduser",
			Email:    "updated@example.com",
		})
		if err != nil {
			log.Printf("Update user failed: %v", err)
			span.RecordError(err)
		} else {
			fmt.Printf("Updated user: %+v\n", updateResp.User)
			span.SetAttributes(attribute.Bool("update.success", true))
		}
		span.End()
	}

	// Test list users
	fmt.Println("\n=== Testing List Users ===")
	_, span = tracer.Start(callCtx, "List Users Test")
	span.SetAttributes(
		attribute.Int("page.size", 10),
		attribute.Int("page.number", 1),
	)
	listResp, err := client.ListUsers(callCtx, &pb.ListUsersRequest{
		PageSize:   10,
		PageNumber: 1,
	})
	if err != nil {
		log.Printf("List users failed: %v", err)
		span.RecordError(err)
	} else {
		fmt.Printf("Listed %d users (total: %d)\n", len(listResp.Users), listResp.TotalCount)
		for i, user := range listResp.Users {
			fmt.Printf("  %d. %+v\n", i+1, user)
		}
		span.SetAttributes(
			attribute.Int("users.count", len(listResp.Users)),
			attribute.Int64("users.total", int64(listResp.TotalCount)),
		)
	}
	span.End()

	fmt.Println("\n=== gRPC Client Test Completed ===")

	// Add final attributes to root span
	rootSpan.SetAttributes(
		attribute.String("client.version", "1.0.0"),
		attribute.String("server.address", "localhost:9090"),
		attribute.Bool("test.completed", true),
	)
}

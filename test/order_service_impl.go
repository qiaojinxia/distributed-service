package test

import (
	"context"
	"fmt"
	"time"

	pb "distributed-service/test/distributed-service/test/proto/order"
)

// OrderServiceImpl 订单服务实现
type OrderServiceImpl struct {
	pb.UnimplementedOrderServiceServer
}

// GetOrder 获取订单信息
func (s *OrderServiceImpl) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	// 模拟处理时间
	time.Sleep(10 * time.Millisecond)

	return &pb.GetOrderResponse{
		Order: &pb.Order{
			Id:          req.OrderId,
			UserId:      "user-123",
			TotalAmount: 99.99,
			CreatedAt:   time.Now().Unix(),
			Items: []*pb.OrderItem{
				{
					ProductId:   "product-1",
					ProductName: "Test Product",
					Quantity:    1,
					Price:       99.99,
				},
			},
		},
	}, nil
}

// CreateOrder 创建订单
func (s *OrderServiceImpl) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	// 模拟处理时间（写操作通常耗时更长）
	time.Sleep(50 * time.Millisecond)

	// 计算总金额
	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	return &pb.CreateOrderResponse{
		Order: &pb.Order{
			Id:          fmt.Sprintf("order-%d", time.Now().Unix()),
			UserId:      req.UserId,
			Items:       req.Items,
			TotalAmount: totalAmount,
			CreatedAt:   time.Now().Unix(),
		},
	}, nil
}

package idgen

import "context"

type IDGenerator interface {
	NextID(ctx context.Context, bizTag string) (int64, error)
	BatchNextID(ctx context.Context, bizTag string, count int) ([]int64, error)
}
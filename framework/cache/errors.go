package cache

import "fmt"

// ErrKeyNotFound 键不存在错误
var ErrKeyNotFound = fmt.Errorf("key not found")

// ErrCacheNotFound 缓存实例不存在错误
var ErrCacheNotFound = fmt.Errorf("cache not found")

// ErrStatsNotSupported 统计功能不支持错误
var ErrStatsNotSupported = fmt.Errorf("stats not supported")

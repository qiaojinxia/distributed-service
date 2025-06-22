package logger

import (
	"time"

	"go.uber.org/zap"
)

// 字段创建工具函数 - 简化日志字段的创建

// String 创建字符串字段
func String(key, value string) Field {
	return zap.String(key, value)
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Int32 创建32位整数字段
func Int32(key string, value int32) Field {
	return zap.Int32(key, value)
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// Uint 创建无符号整数字段
func Uint(key string, value uint) Field {
	return zap.Uint(key, value)
}

// Uint32 创建32位无符号整数字段
func Uint32(key string, value uint32) Field {
	return zap.Uint32(key, value)
}

// Uint64 创建64位无符号整数字段
func Uint64(key string, value uint64) Field {
	return zap.Uint64(key, value)
}

// Float32 创建32位浮点字段
func Float32(key string, value float32) Field {
	return zap.Float32(key, value)
}

// Float64 创建64位浮点字段
func Float64(key string, value float64) Field {
	return zap.Float64(key, value)
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

// Duration 创建时间间隔字段
func Duration(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

// Time 创建时间字段
func Time(key string, value time.Time) Field {
	return zap.Time(key, value)
}

// Err  创建错误字段
func Err(err error) Field {
	return zap.Error(err)
}

// Errors 创建错误字段的别名（兼容性）
func Errors(err error) Field {
	return zap.Error(err)
}

// Error_ 创建错误字段的旧API兼容函数
func Error_(err error) Field {
	return zap.Error(err)
}

// Any 创建任意类型字段（性能较低，谨慎使用）
func Any(key string, value interface{}) Field {
	return zap.Any(key, value)
}

// ByteString 创建字节字符串字段
func ByteString(key string, value []byte) Field {
	return zap.ByteString(key, value)
}

// Complex64 创建64位复数字段
func Complex64(key string, value complex64) Field {
	return zap.Complex64(key, value)
}

// Complex128 创建128位复数字段
func Complex128(key string, value complex128) Field {
	return zap.Complex128(key, value)
}

// Strings 创建字符串数组字段
func Strings(key string, values []string) Field {
	return zap.Strings(key, values)
}

// Ints 创建整数数组字段
func Ints(key string, values []int) Field {
	return zap.Ints(key, values)
}

// 业务相关的便捷字段创建函数

// UserID 创建用户ID字段
func UserID(id string) Field {
	return String("user_id", id)
}

// RequestID 创建请求ID字段
func RequestID(id string) Field {
	return String("request_id", id)
}

// TraceID 创建链路追踪ID字段
func TraceID(id string) Field {
	return String("trace_id", id)
}

// SpanID 创建Span ID字段
func SpanID(id string) Field {
	return String("span_id", id)
}

// Method 创建HTTP方法字段
func Method(method string) Field {
	return String("method", method)
}

// Path 创建请求路径字段
func Path(path string) Field {
	return String("path", path)
}

// StatusCode 创建状态码字段
func StatusCode(code int) Field {
	return Int("status_code", code)
}

// ClientIP 创建客户端IP字段
func ClientIP(ip string) Field {
	return String("client_ip", ip)
}

// Operation 创建操作字段
func Operation(op string) Field {
	return String("operation", op)
}

// Service 创建服务名字段
func Service(name string) Field {
	return String("service", name)
}

// Version 创建版本字段
func Version(version string) Field {
	return String("version", version)
}

// Environment 创建环境字段
func Environment(env string) Field {
	return String("environment", env)
}

// Database 创建数据库名字段
func Database(name string) Field {
	return String("database", name)
}

// Table 创建表名字段
func Table(name string) Field {
	return String("table", name)
}

// SQL 创建SQL语句字段
func SQL(query string) Field {
	return String("sql", query)
}

// Queue 创建队列名字段
func Queue(name string) Field {
	return String("queue", name)
}

// Topic 创建主题字段
func Topic(name string) Field {
	return String("topic", name)
}

// 性能相关字段

// Latency 创建延迟字段
func Latency(d time.Duration) Field {
	return Duration("latency", d)
}

// ResponseTime 创建响应时间字段
func ResponseTime(d time.Duration) Field {
	return Duration("response_time", d)
}

// ResponseSize 创建响应大小字段
func ResponseSize(size int) Field {
	return Int("response_size", size)
}

// Count 创建计数字段
func Count(count int) Field {
	return Int("count", count)
}

// Size 创建大小字段
func Size(size int) Field {
	return Int("size", size)
}

// Fields 批量字段创建器
type Fields struct {
	fields []Field
}

// NewFields 创建新的字段构建器
func NewFields() *Fields {
	return &Fields{fields: make([]Field, 0)}
}

// Add 添加字段
func (f *Fields) Add(field Field) *Fields {
	f.fields = append(f.fields, field)
	return f
}

// String 添加字符串字段
func (f *Fields) String(key, value string) *Fields {
	return f.Add(String(key, value))
}

// Int 添加整数字段
func (f *Fields) Int(key string, value int) *Fields {
	return f.Add(Int(key, value))
}

// Bool 添加布尔字段
func (f *Fields) Bool(key string, value bool) *Fields {
	return f.Add(Bool(key, value))
}

// Error 添加错误字段
func (f *Fields) Error(err error) *Fields {
	return f.Add(Err(err))
}

// Duration 添加时间间隔字段
func (f *Fields) Duration(key string, value time.Duration) *Fields {
	return f.Add(Duration(key, value))
}

// Build 构建字段数组
func (f *Fields) Build() []Field {
	return f.fields
}

// 便捷的链式调用示例：
// fields := logger.NewFields().
//     String("user_id", "123").
//     Int("status_code", 200).
//     Duration("latency", time.Millisecond*100).
//     Build()

// Port 创建端口字段
func Port(port int) Field {
	return Int("port", port)
}

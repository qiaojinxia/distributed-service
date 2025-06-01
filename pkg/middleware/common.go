// Package middleware provides common utilities for middleware
package middleware

import (
	"strings"
)

// ===== String Conversion Helpers =====

// convertToCamelCase 将下划线命名转换为驼峰命名
// create_user -> CreateUser
func convertToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// convertToSnakeCase 将驼峰命名转换为下划线命名
// CreateUser -> create_user
func convertToSnakeCase(s string) string {
	var result strings.Builder
	for i, char := range s {
		if i > 0 && 'A' <= char && char <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(rune(strings.ToLower(string(char))[0]))
	}
	return result.String()
}

// isUpperCamelCase 检查字符串是否为大驼峰命名格式
func isUpperCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// 首字母大写且不包含下划线
	return 'A' <= s[0] && s[0] <= 'Z' && !strings.Contains(s, "_")
}

// ===== Helper Functions for gRPC =====

// findLastSlash finds the last slash in a string
func findLastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}

// findLastDot finds the last dot in a string
func findLastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

// getServiceName extracts service name from full method name
// /userpb.UserService/CreateUser -> UserService
func getServiceName(fullMethod string) string {
	if idx := findLastSlash(fullMethod); idx >= 0 {
		if serviceMethod := fullMethod[1:idx]; len(serviceMethod) > 0 {
			if dotIdx := findLastDot(serviceMethod); dotIdx >= 0 {
				return serviceMethod[dotIdx+1:]
			}
			return serviceMethod
		}
	}
	return "unknown"
}

// getMethodName extracts method name from full method name
// /userpb.UserService/CreateUser -> CreateUser
func getMethodName(fullMethod string) string {
	if idx := findLastSlash(fullMethod); idx >= 0 && idx < len(fullMethod)-1 {
		return fullMethod[idx+1:]
	}
	return "unknown"
}

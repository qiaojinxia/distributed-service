package plugin

import (
	"fmt"
	"reflect"
	"sync"
)

// DefaultRegistry 默认插件注册表实现
type DefaultRegistry struct {
	plugins   map[string]Plugin
	typeIndex map[string][]string // 类型索引，用于快速查找特定类型的插件
	mu        sync.RWMutex
}

// NewDefaultRegistry 创建默认注册表
func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		plugins:   make(map[string]Plugin),
		typeIndex: make(map[string][]string),
	}
}

// Register 注册插件
func (r *DefaultRegistry) Register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否已存在
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' already registered", name)
	}

	// 注册插件
	r.plugins[name] = plugin

	// 更新类型索引
	pluginType := r.getPluginType(plugin)
	if pluginType != "" {
		r.typeIndex[pluginType] = append(r.typeIndex[pluginType], name)
	}

	return nil
}

// Unregister 注销插件
func (r *DefaultRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// 从插件映射中移除
	delete(r.plugins, name)

	// 从类型索引中移除
	pluginType := r.getPluginType(plugin)
	if pluginType != "" {
		if plugins, exists := r.typeIndex[pluginType]; exists {
			for i, pluginName := range plugins {
				if pluginName == name {
					r.typeIndex[pluginType] = append(plugins[:i], plugins[i+1:]...)
					break
				}
			}
			// 如果类型下没有插件了，移除类型索引
			if len(r.typeIndex[pluginType]) == 0 {
				delete(r.typeIndex, pluginType)
			}
		}
	}

	return nil
}

// Get 获取插件
func (r *DefaultRegistry) Get(name string) Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.plugins[name]
}

// GetAll 获取所有插件
func (r *DefaultRegistry) GetAll() map[string]Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Plugin)
	for name, plugin := range r.plugins {
		result[name] = plugin
	}
	return result
}

// GetByType 根据类型获取插件
func (r *DefaultRegistry) GetByType(pluginType string) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Plugin
	if pluginNames, exists := r.typeIndex[pluginType]; exists {
		for _, name := range pluginNames {
			if plugin, exists := r.plugins[name]; exists {
				result = append(result, plugin)
			}
		}
	}
	return result
}

// Exists 检查插件是否存在
func (r *DefaultRegistry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[name]
	return exists
}

// GetTypes 获取所有已注册的插件类型
func (r *DefaultRegistry) GetTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var types []string
	for pluginType := range r.typeIndex {
		types = append(types, pluginType)
	}
	return types
}

// GetPluginsByInterface 根据接口类型获取插件
func (r *DefaultRegistry) GetPluginsByInterface(interfaceType interface{}) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Plugin
	targetType := reflect.TypeOf(interfaceType).Elem()

	for _, plugin := range r.plugins {
		pluginType := reflect.TypeOf(plugin)
		if pluginType.Implements(targetType) {
			result = append(result, plugin)
		}
	}
	return result
}

// Count 获取插件数量
func (r *DefaultRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.plugins)
}

// CountByType 根据类型获取插件数量
func (r *DefaultRegistry) CountByType(pluginType string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if pluginNames, exists := r.typeIndex[pluginType]; exists {
		return len(pluginNames)
	}
	return 0
}

// Clear 清空所有插件
func (r *DefaultRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.plugins = make(map[string]Plugin)
	r.typeIndex = make(map[string][]string)
}

// getPluginType 获取插件类型
func (r *DefaultRegistry) getPluginType(plugin Plugin) string {
	switch plugin.(type) {
	case ServicePlugin:
		return "service"
	case MiddlewarePlugin:
		return "middleware"
	case TransportPlugin:
		return "transport"
	default:
		return "generic"
	}
}

// ValidatePlugin 验证插件
func (r *DefaultRegistry) ValidatePlugin(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	version := plugin.Version()
	if version == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	// 验证依赖项
	dependencies := plugin.Dependencies()
	for _, dep := range dependencies {
		if dep == name {
			return fmt.Errorf("plugin cannot depend on itself")
		}
	}

	return nil
}

// GetDependents 获取依赖于指定插件的插件列表
func (r *DefaultRegistry) GetDependents(pluginName string) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var dependents []Plugin
	for _, plugin := range r.plugins {
		dependencies := plugin.Dependencies()
		for _, dep := range dependencies {
			if dep == pluginName {
				dependents = append(dependents, plugin)
				break
			}
		}
	}
	return dependents
}

// GetDependencies 获取指定插件的所有依赖
func (r *DefaultRegistry) GetDependencies(pluginName string) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[pluginName]
	if !exists {
		return nil
	}

	var dependencies []Plugin
	for _, depName := range plugin.Dependencies() {
		if depPlugin, exists := r.plugins[depName]; exists {
			dependencies = append(dependencies, depPlugin)
		}
	}
	return dependencies
}

// HasCircularDependency 检查是否存在循环依赖
func (r *DefaultRegistry) HasCircularDependency() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for name := range r.plugins {
		if !visited[name] {
			if r.hasCycleDFS(name, visited, recStack) {
				return true
			}
		}
	}
	return false
}

// hasCycleDFS 深度优先搜索检查循环依赖
func (r *DefaultRegistry) hasCycleDFS(pluginName string, visited, recStack map[string]bool) bool {
	visited[pluginName] = true
	recStack[pluginName] = true

	plugin, exists := r.plugins[pluginName]
	if !exists {
		return false
	}

	for _, depName := range plugin.Dependencies() {
		if !visited[depName] {
			if r.hasCycleDFS(depName, visited, recStack) {
				return true
			}
		} else if recStack[depName] {
			return true
		}
	}

	recStack[pluginName] = false
	return false
}

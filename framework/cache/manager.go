package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	caches   map[string]Cache
	builders map[Type]Builder
	mutex    sync.RWMutex
}

type Builder interface {
	Build(config Config) (Cache, error)
}

type Config struct {
	Type     Type                   `json:"type" yaml:"type"`
	Name     string                 `json:"name" yaml:"name"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

func NewManager() *Manager {
	return &Manager{
		caches:   make(map[string]Cache),
		builders: make(map[Type]Builder),
	}
}

func (m *Manager) RegisterBuilder(cacheType Type, builder Builder) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.builders[cacheType] = builder
}

func (m *Manager) CreateCache(config Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.caches[config.Name]; exists {
		return fmt.Errorf("cache %s already exists", config.Name)
	}

	builder, exists := m.builders[config.Type]
	if !exists {
		return fmt.Errorf("no builder registered for cache type %s", config.Type)
	}

	cache, err := builder.Build(config)
	if err != nil {
		return fmt.Errorf("failed to build cache %s: %w", config.Name, err)
	}

	m.caches[config.Name] = cache
	return nil
}

func (m *Manager) GetCache(name string) (Cache, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cache, exists := m.caches[name]
	if !exists {
		return nil, fmt.Errorf("cache %s not found", name)
	}

	return cache, nil
}

func (m *Manager) RemoveCache(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cache, exists := m.caches[name]
	if !exists {
		return fmt.Errorf("cache %s not found", name)
	}

	if err := cache.Close(); err != nil {
		return fmt.Errorf("failed to close cache %s: %w", name, err)
	}

	delete(m.caches, name)
	return nil
}

func (m *Manager) ListCaches() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.caches))
	for name := range m.caches {
		names = append(names, name)
	}
	return names
}

func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var lastErr error
	for name, cache := range m.caches {
		if err := cache.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close cache %s: %w", name, err)
		}
	}

	m.caches = make(map[string]Cache)
	return lastErr
}

type Wrapper struct {
	cache Cache
	name  string
}

func (w *Wrapper) Get(ctx context.Context, key string) (interface{}, error) {
	return w.cache.Get(ctx, w.prefixKey(key))
}

func (w *Wrapper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return w.cache.Set(ctx, w.prefixKey(key), value, expiration)
}

func (w *Wrapper) Delete(ctx context.Context, key string) error {
	return w.cache.Delete(ctx, w.prefixKey(key))
}

func (w *Wrapper) Exists(ctx context.Context, key string) (bool, error) {
	return w.cache.Exists(ctx, w.prefixKey(key))
}

func (w *Wrapper) Clear(ctx context.Context) error {
	return w.cache.Clear(ctx)
}

func (w *Wrapper) Close() error {
	return w.cache.Close()
}

func (w *Wrapper) prefixKey(key string) string {
	return fmt.Sprintf("%s:%s", w.name, key)
}

func (m *Manager) GetNamedCache(name string) Cache {
	cache, err := m.GetCache(name)
	if err != nil {
		return nil
	}
	return &Wrapper{cache: cache, name: name}
}

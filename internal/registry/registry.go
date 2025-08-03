package registry

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Tags     []string          `json:"tags"`
	Meta     map[string]string `json:"meta"`
	Healthy  bool              `json:"healthy"`
	LastSeen time.Time         `json:"last_seen"`
}

// Registry 服务注册中心接口
type Registry interface {
	Register(ctx context.Context, service *ServiceInfo) error
	Deregister(ctx context.Context, serviceID string) error
	Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error)
	HealthCheck(ctx context.Context, serviceID string) error
	Close() error
}

// ConsulRegistry Consul注册中心实现
type ConsulRegistry struct {
	client *consulapi.Client
	config *consulapi.Config
}

// NewConsulRegistry 创建Consul注册中心
func NewConsulRegistry(address string) (*ConsulRegistry, error) {
	config := consulapi.DefaultConfig()
	if address != "" {
		config.Address = address
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 测试连接
	_, err = client.Status().Leader()
	if err != nil {
		log.Printf("Warning: Consul connection failed: %v, falling back to memory registry", err)
		return nil, err
	}

	return &ConsulRegistry{
		client: client,
		config: config,
	}, nil
}

// Register 注册服务
func (c *ConsulRegistry) Register(ctx context.Context, service *ServiceInfo) error {
	registration := &consulapi.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Meta:    service.Meta,
	}

	// 添加健康检查
	registration.Check = &consulapi.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
		Interval:                       "10s",
		Timeout:                        "5s",
		DeregisterCriticalServiceAfter: "30s",
	}

	return c.client.Agent().ServiceRegister(registration)
}

// Deregister 注销服务
func (c *ConsulRegistry) Deregister(ctx context.Context, serviceID string) error {
	return c.client.Agent().ServiceDeregister(serviceID)
}

// Discover 发现服务
func (c *ConsulRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	var result []*ServiceInfo
	for _, service := range services {
		result = append(result, &ServiceInfo{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
			Meta:    service.Service.Meta,
			Healthy: len(service.Checks) > 0,
		})
	}

	return result, nil
}

// HealthCheck 健康检查
func (c *ConsulRegistry) HealthCheck(ctx context.Context, serviceID string) error {
	// Consul会自动进行健康检查
	return nil
}

// Close 关闭连接
func (c *ConsulRegistry) Close() error {
	return nil
}

// MemoryRegistry 内存注册中心实现（备用方案）
type MemoryRegistry struct {
	services map[string]*ServiceInfo
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewMemoryRegistry 创建内存注册中心
func NewMemoryRegistry() *MemoryRegistry {
	ctx, cancel := context.WithCancel(context.Background())
	r := &MemoryRegistry{
		services: make(map[string]*ServiceInfo),
		ctx:      ctx,
		cancel:   cancel,
	}

	// 启动健康检查协程
	go r.healthCheckLoop()

	return r
}

// Register 注册服务
func (m *MemoryRegistry) Register(ctx context.Context, service *ServiceInfo) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	service.Healthy = true
	service.LastSeen = time.Now()
	m.services[service.ID] = service

	log.Printf("Service registered: %s (%s:%d)", service.Name, service.Address, service.Port)
	return nil
}

// Deregister 注销服务
func (m *MemoryRegistry) Deregister(ctx context.Context, serviceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if service, exists := m.services[serviceID]; exists {
		delete(m.services, serviceID)
		log.Printf("Service deregistered: %s", service.Name)
	}

	return nil
}

// Discover 发现服务
func (m *MemoryRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []*ServiceInfo
	for _, service := range m.services {
		if service.Name == serviceName && service.Healthy {
			result = append(result, service)
		}
	}

	return result, nil
}

// HealthCheck 健康检查
func (m *MemoryRegistry) HealthCheck(ctx context.Context, serviceID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if service, exists := m.services[serviceID]; exists {
		service.LastSeen = time.Now()
		service.Healthy = true
	}

	return nil
}

// Close 关闭注册中心
func (m *MemoryRegistry) Close() error {
	m.cancel()
	return nil
}

// healthCheckLoop 健康检查循环
func (m *MemoryRegistry) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.mutex.Lock()
			now := time.Now()
			for id, service := range m.services {
				// 如果服务超过60秒没有心跳，标记为不健康
				if now.Sub(service.LastSeen) > 60*time.Second {
					service.Healthy = false
					log.Printf("Service marked as unhealthy: %s", service.Name)
				}
				// 如果服务超过120秒没有心跳，移除服务
				if now.Sub(service.LastSeen) > 120*time.Second {
					delete(m.services, id)
					log.Printf("Service removed due to timeout: %s", service.Name)
				}
			}
			m.mutex.Unlock()
		}
	}
}

// NewRegistry 创建注册中心（优先使用Consul，失败则使用内存注册中心）
func NewRegistry(consulAddress string) Registry {
	// 尝试连接Consul
	if consulRegistry, err := NewConsulRegistry(consulAddress); err == nil {
		log.Println("Using Consul registry")
		return consulRegistry
	}

	// 回退到内存注册中心
	log.Println("Using memory registry as fallback")
	return NewMemoryRegistry()
}
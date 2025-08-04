package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"
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

// EtcdRegistry etcd注册中心实现
type EtcdRegistry struct {
	client   *clientv3.Client
	leaseID  clientv3.LeaseID
	services map[string]*ServiceInfo
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewEtcdRegistry 创建etcd注册中心
func NewEtcdRegistry(endpoints []string) (*EtcdRegistry, error) {
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &EtcdRegistry{
		client:   client,
		services: make(map[string]*ServiceInfo),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
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

// Register 注册服务到etcd
func (e *EtcdRegistry) Register(ctx context.Context, service *ServiceInfo) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 创建租约
	lease, err := e.client.Grant(ctx, 30) // 30秒租约
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}
	e.leaseID = lease.ID

	// 序列化服务信息
	serviceData, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service: %w", err)
	}

	// 构建key
	key := fmt.Sprintf("/services/%s/%s", service.Name, service.ID)

	// 注册服务
	_, err = e.client.Put(ctx, key, string(serviceData), clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 保持租约
	ch, kaerr := e.client.KeepAlive(e.ctx, lease.ID)
	if kaerr != nil {
		return fmt.Errorf("failed to keep alive lease: %w", kaerr)
	}

	// 启动租约续期协程
	go func() {
		for ka := range ch {
			_ = ka // 忽略续期响应
		}
	}()

	// 保存服务信息
	e.services[service.ID] = service

	return nil
}

// Deregister 从etcd注销服务
func (e *EtcdRegistry) Deregister(ctx context.Context, serviceID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	service, exists := e.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// 构建key
	key := fmt.Sprintf("/services/%s/%s", service.Name, serviceID)

	// 删除服务
	_, err := e.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	// 从本地缓存删除
	delete(e.services, serviceID)

	return nil
}

// Discover 从etcd发现服务
func (e *EtcdRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInfo, error) {
	key := fmt.Sprintf("/services/%s/", serviceName)

	resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	var services []*ServiceInfo
	for _, kv := range resp.Kvs {
		var service ServiceInfo
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			log.Printf("Failed to unmarshal service data: %v", err)
			continue
		}
		services = append(services, &service)
	}

	return services, nil
}

// HealthCheck etcd健康检查
func (e *EtcdRegistry) HealthCheck(ctx context.Context, serviceID string) error {
	// etcd通过租约机制自动处理健康检查
	// 这里可以检查服务是否仍然存在
	service, exists := e.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	key := fmt.Sprintf("/services/%s/%s", service.Name, serviceID)
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check service health: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("service %s not found in etcd", serviceID)
	}

	return nil
}

// Close 关闭etcd连接
func (e *EtcdRegistry) Close() error {
	e.cancel()
	return e.client.Close()
}

// MemoryRegistry 内存注册中心实现（备用方案）
type MemoryRegistry struct {
	services map[string]*ServiceInfo
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

var (
	memoryRegistryInstance *MemoryRegistry
	memoryRegistryOnce     sync.Once
)

// NewMemoryRegistry 创建内存注册中心（单例模式）
func NewMemoryRegistry() *MemoryRegistry {
	memoryRegistryOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		memoryRegistryInstance = &MemoryRegistry{
			services: make(map[string]*ServiceInfo),
			ctx:      ctx,
			cancel:   cancel,
		}

		// 启动健康检查协程
		go memoryRegistryInstance.healthCheckLoop()
	})

	return memoryRegistryInstance
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

// NewRegistry 创建注册中心实例
func NewRegistry(registryType, address string) Registry {
	switch strings.ToLower(registryType) {
	case "etcd":
		var endpoints []string
		if address != "" {
			endpoints = strings.Split(address, ",")
		}
		etcdRegistry, err := NewEtcdRegistry(endpoints)
		if err != nil {
			log.Printf("Failed to create etcd registry: %v, falling back to memory registry", err)
			return NewMemoryRegistry()
		}
		return etcdRegistry
	case "consul":
		consulRegistry, err := NewConsulRegistry(address)
		if err != nil {
			log.Printf("Failed to create consul registry: %v, falling back to memory registry", err)
			return NewMemoryRegistry()
		}
		return consulRegistry
	default:
		return NewMemoryRegistry()
	}
}
package tests

import (
	"context"
	"testing"
	"time"

	"chain/internal/registry"
)

func TestServiceDiscovery(t *testing.T) {
	// 创建注册中心（使用与服务器相同的方式）
	reg := registry.NewRegistry("memory", "")

	// 模拟注册一个服务
	service := &registry.ServiceInfo{
		ID:      "test-service-1",
		Name:    "test-service",
		Address: "localhost",
		Port:    8080,
		Tags:    []string{"test"},
		Meta:    map[string]string{"version": "1.0"},
	}

	err := reg.Register(context.Background(), service)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// 发现服务
	services, err := reg.Discover(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("Failed to discover service: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(services))
	}

	if services[0].Name != "test-service" {
		t.Fatalf("Expected service name 'test-service', got '%s'", services[0].Name)
	}

	// 注销服务
	err = reg.Deregister(context.Background(), "test-service-1")
	if err != nil {
		t.Fatalf("Failed to deregister service: %v", err)
	}

	// 再次发现服务，应该为空
	services, err = reg.Discover(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("Failed to discover service after deregistration: %v", err)
	}

	if len(services) != 0 {
		t.Fatalf("Expected 0 services after deregistration, got %d", len(services))
	}
}

func TestChainGrpcServiceDiscovery(t *testing.T) {
	// 创建注册中心
	reg := registry.NewRegistry("memory", "")

	// 等待一段时间，让服务器有机会注册
	time.Sleep(1 * time.Second)

	// 发现 chain-grpc 服务
	services, err := reg.Discover(context.Background(), "chain-grpc")
	if err != nil {
		t.Fatalf("Failed to discover chain-grpc service: %v", err)
	}

	t.Logf("Discovered %d chain-grpc services:", len(services))
	for _, service := range services {
		t.Logf("Service: %s, Address: %s, Port: %d", service.Name, service.Address, service.Port)
	}
}

func TestHealthCheck(t *testing.T) {
	// 创建注册中心
	reg := registry.NewRegistry("memory", "")

	// 注册一个服务
	service := &registry.ServiceInfo{
		ID:      "health-test-service",
		Name:    "health-test",
		Address: "localhost",
		Port:    9999,
	}

	err := reg.Register(context.Background(), service)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// 执行健康检查
	err = reg.HealthCheck(context.Background(), "health-test-service")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	// 清理
	err = reg.Deregister(context.Background(), "health-test-service")
	if err != nil {
		t.Fatalf("Failed to deregister service: %v", err)
	}
}
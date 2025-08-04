package tests

import (
	"context"
	"testing"
	"time"

	"chain/internal/registry"
)

func TestEtcdServiceDiscovery(t *testing.T) {
	// 创建 etcd 注册中心
	reg := registry.NewRegistry("etcd", "localhost:2379")
	if reg == nil {
		t.Fatal("Failed to create etcd registry")
	}
	defer reg.Close()

	// 等待一段时间，让服务器有机会注册
	time.Sleep(2 * time.Second)

	// 发现 chain-grpc 服务
	services, err := reg.Discover(context.Background(), "chain-grpc")
	if err != nil {
		t.Fatalf("Failed to discover chain-grpc service: %v", err)
	}

	t.Logf("Discovered %d chain-grpc services in etcd:", len(services))
	for _, service := range services {
		t.Logf("Service: %s, ID: %s, Address: %s, Port: %d", service.Name, service.ID, service.Address, service.Port)
	}

	if len(services) == 0 {
		t.Log("No services found, this might be expected if the server is not running")
	} else {
		// 验证服务信息
		service := services[0]
		if service.Name != "chain-grpc" {
			t.Errorf("Expected service name 'chain-grpc', got '%s'", service.Name)
		}
		if service.Port != 9090 {
			t.Errorf("Expected port 9090, got %d", service.Port)
		}
	}
}

func TestEtcdServiceRegistration(t *testing.T) {
	// 创建 etcd 注册中心
	reg := registry.NewRegistry("etcd", "localhost:2379")
	if reg == nil {
		t.Fatal("Failed to create etcd registry")
	}
	defer reg.Close()

	// 注册一个测试服务
	service := &registry.ServiceInfo{
		ID:      "test-etcd-service-1",
		Name:    "test-etcd-service",
		Address: "localhost",
		Port:    8888,
		Tags:    []string{"test", "etcd"},
		Meta:    map[string]string{"version": "1.0", "env": "test"},
	}

	err := reg.Register(context.Background(), service)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// 等待注册完成
	time.Sleep(1 * time.Second)

	// 发现刚注册的服务
	services, err := reg.Discover(context.Background(), "test-etcd-service")
	if err != nil {
		t.Fatalf("Failed to discover service: %v", err)
	}

	if len(services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(services))
	}

	foundService := services[0]
	if foundService.Name != "test-etcd-service" {
		t.Errorf("Expected service name 'test-etcd-service', got '%s'", foundService.Name)
	}
	if foundService.Port != 8888 {
		t.Errorf("Expected port 8888, got %d", foundService.Port)
	}

	// 注销服务
	err = reg.Deregister(context.Background(), "test-etcd-service-1")
	if err != nil {
		t.Fatalf("Failed to deregister service: %v", err)
	}

	// 等待注销完成
	time.Sleep(1 * time.Second)

	// 再次发现服务，应该为空
	services, err = reg.Discover(context.Background(), "test-etcd-service")
	if err != nil {
		t.Fatalf("Failed to discover service after deregistration: %v", err)
	}

	if len(services) != 0 {
		t.Fatalf("Expected 0 services after deregistration, got %d", len(services))
	}

	t.Log("Etcd service registration and deregistration test passed")
}
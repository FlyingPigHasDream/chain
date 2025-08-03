package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"chain/internal/config"
	grpcServer "chain/internal/grpc"
	"chain/pkg/logger"
)

func main() {
	// 初始化日志
	logger.Init("info")

	// 加载配置
	cfg := config.Load()

	// 创建gRPC服务器
	server := grpcServer.NewServer(cfg)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 等待中断信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// 优雅关闭
	log.Println("Shutting down gRPC server...")
	server.Stop()
	log.Println("gRPC server stopped")
}

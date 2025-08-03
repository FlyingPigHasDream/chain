package main

import (
	"log"
	"os"

	"chain/internal/config"
	"chain/internal/server"
	"chain/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 初始化配置
	cfg := config.Load()

	// 初始化日志
	logger.Init(cfg.LogLevel)

	// 启动服务器
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
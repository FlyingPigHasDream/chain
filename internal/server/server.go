package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chain/internal/config"
	"chain/internal/database"
	"chain/internal/handlers"
	"chain/internal/models"
	"chain/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务器
type Server struct {
	config *config.Config
	router *gin.Engine
	server *http.Server
	db     *database.Database
}

// New 创建新的服务器实例
func New(cfg *config.Config) *Server {
	// 设置Gin模式
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		panic(err)
	}

	// 自动迁移数据库表
	err = db.AutoMigrate(
		&models.Transaction{},
		&models.Block{},
		&models.Account{},
		&models.Token{},
		&models.TokenBalance{},
	)
	if err != nil {
		logger.Error("Failed to migrate database: %v", err)
		panic(err)
	}

	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 注册路由
	handlers.RegisterRoutes(router, cfg, db)

	return &Server{
		config: cfg,
		router: router,
		db:     db,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)
	
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// 启动服务器
	go func() {
		logger.Infof("Server starting on %s", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	logger.Info("Server exited")
	return nil
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
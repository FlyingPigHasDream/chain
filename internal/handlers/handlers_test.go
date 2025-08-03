package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chain/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建路由
	router := gin.New()
	router.GET("/health", healthCheck)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// 执行请求
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestRegisterRoutes(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Host: "localhost",
		},
		Chain: config.ChainConfig{
			RPCURL:     "https://mainnet.infura.io/v3/test",
			PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ChainID:    1,
			GasLimit:   21000,
		},
		LogLevel: "info",
	}

	// 创建路由
	router := gin.New()
	RegisterRoutes(router, cfg)

	// 测试健康检查路由
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetBalanceInvalidAddress(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		Chain: config.ChainConfig{
			RPCURL:     "https://mainnet.infura.io/v3/test",
			PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ChainID:    1,
			GasLimit:   21000,
		},
	}

	// 创建处理器
	handler := NewChainHandler(cfg)

	// 创建路由
	router := gin.New()
	router.GET("/balance/:address", handler.GetBalance)

	// 测试空地址
	req, _ := http.NewRequest("GET", "/balance/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// 由于路由不匹配，应该返回404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTransferInvalidRequest(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		Chain: config.ChainConfig{
			RPCURL:     "https://mainnet.infura.io/v3/test",
			PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ChainID:    1,
			GasLimit:   21000,
		},
	}

	// 创建处理器
	handler := NewChainHandler(cfg)

	// 创建路由
	router := gin.New()
	router.POST("/transfer", handler.Transfer)

	// 测试无效的JSON
	invalidJSON := `{"to": "0x123"}` // 缺少amount字段
	req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 应该返回400错误
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
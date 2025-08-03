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

func TestBSCRoutes(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Host: "localhost",
		},
		Chain: config.ChainConfig{
			RPCURL:     "https://bsc-dataseed1.binance.org/",
			PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ChainID:    56,
			GasLimit:   21000,
		},
		LogLevel: "info",
	}

	// 创建路由
	router := gin.New()
	RegisterBSCRoutes(router, cfg)

	// 测试代币搜索路由
	req, _ := http.NewRequest("GET", "/api/v1/bsc/token/search/WBNB", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// 由于需要实际的网络连接，这里只测试路由是否正确注册
	// 在实际环境中，这个测试可能会失败，因为需要真实的BSC连接
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestBSCTokenPriceRequest(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		Chain: config.ChainConfig{
			RPCURL:     "https://bsc-dataseed1.binance.org/",
			PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ChainID:    56,
			GasLimit:   21000,
		},
	}

	// 创建处理器
	handler := NewBSCHandler(cfg)

	// 创建路由
	router := gin.New()
	router.POST("/token/price", handler.GetTokenPriceByAddressAndName)

	// 测试有效请求
	validRequest := map[string]interface{}{
		"address":    "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c", // WBNB
		"token_name": "WBNB",
	}
	jsonData, _ := json.Marshal(validRequest)
	req, _ := http.NewRequest("POST", "/token/price", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 由于需要实际的网络连接，这里只测试请求格式是否正确
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestBSCInvalidAddressFormat(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		Chain: config.ChainConfig{
			RPCURL:     "https://bsc-dataseed1.binance.org/",
			PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ChainID:    56,
			GasLimit:   21000,
		},
	}

	// 创建处理器
	handler := NewBSCHandler(cfg)

	// 创建路由
	router := gin.New()
	router.GET("/token/info/:address", handler.GetTokenInfo)

	// 测试无效地址格式
	req, _ := http.NewRequest("GET", "/token/info/invalid-address", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回400错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid token address format")
}

func TestBSCMultipleTokenPricesValidation(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		Chain: config.ChainConfig{
			RPCURL:     "https://bsc-dataseed1.binance.org/",
			PrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ChainID:    56,
			GasLimit:   21000,
		},
	}

	// 创建处理器
	handler := NewBSCHandler(cfg)

	// 创建路由
	router := gin.New()
	router.POST("/tokens/prices", handler.GetMultipleTokenPrices)

	// 测试超过限制的代币数量
	tokens := make([]map[string]string, 11) // 超过10个限制
	for i := 0; i < 11; i++ {
		tokens[i] = map[string]string{
			"address": "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c",
		}
	}

	requestData := map[string]interface{}{
		"tokens": tokens,
	}
	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/tokens/prices", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 应该返回400错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "maximum 10 tokens allowed")
}
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"chain/internal/config"
	"chain/internal/services"
	"chain/pkg/logger"

	"github.com/gin-gonic/gin"
)

// BSCHandler BSC链处理器
type BSCHandler struct {
	bscService *services.BSCService
}

// NewBSCHandler 创建新的BSC处理器
func NewBSCHandler(cfg *config.Config) *BSCHandler {
	bscService := services.NewBSCService(cfg)
	return &BSCHandler{
		bscService: bscService,
	}
}

// RegisterBSCRoutes 注册BSC相关路由
func RegisterBSCRoutes(router *gin.Engine, cfg *config.Config) {
	bscHandler := NewBSCHandler(cfg)

	// BSC API路由组
	bsc := router.Group("/api/v1/bsc")
	{
		// 代币信息查询
		bsc.GET("/token/info/:address", bscHandler.GetTokenInfo)
		
		// 代币价格查询（通过合约地址）
		bsc.GET("/token/price/:address", bscHandler.GetTokenPrice)
		
		// 代币价格查询（通过合约地址和名称验证）
		bsc.POST("/token/price", bscHandler.GetTokenPriceByAddressAndName)
		
		// 通过名称查找代币
		bsc.GET("/token/search/:name", bscHandler.FindTokenByName)
		
		// 批量查询代币价格
		bsc.POST("/tokens/prices", bscHandler.GetMultipleTokenPrices)
		
		// 获取流动性池信息
		bsc.GET("/liquidity/:tokenA/:tokenB", bscHandler.GetLiquidityInfo)
	}
}

// GetTokenInfo 获取代币信息
func (h *BSCHandler) GetTokenInfo(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token address is required"})
		return
	}

	// 验证地址格式
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token address format"})
		return
	}

	tokenInfo, err := h.bscService.GetTokenInfo(address)
	if err != nil {
		logger.Errorf("Failed to get token info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tokenInfo,
	})
}

// GetTokenPrice 获取代币价格
func (h *BSCHandler) GetTokenPrice(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token address is required"})
		return
	}

	// 验证地址格式
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token address format"})
		return
	}

	priceInfo, err := h.bscService.GetTokenPrice(address, "")
	if err != nil {
		logger.Errorf("Failed to get token price: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    priceInfo,
	})
}

// GetTokenPriceByAddressAndName 通过合约地址和名称获取代币价格
func (h *BSCHandler) GetTokenPriceByAddressAndName(c *gin.Context) {
	var req struct {
		Address   string `json:"address" binding:"required"`
		TokenName string `json:"token_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证地址格式
	if !strings.HasPrefix(req.Address, "0x") || len(req.Address) != 42 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token address format"})
		return
	}

	priceInfo, err := h.bscService.GetTokenPrice(req.Address, req.TokenName)
	if err != nil {
		logger.Errorf("Failed to get token price: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    priceInfo,
	})
}

// FindTokenByName 通过名称查找代币
func (h *BSCHandler) FindTokenByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token name is required"})
		return
	}

	tokens, err := h.bscService.FindTokenByName(name)
	if err != nil {
		logger.Errorf("Failed to find tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tokens,
		"count":   len(tokens),
	})
}

// GetMultipleTokenPrices 批量获取代币价格
func (h *BSCHandler) GetMultipleTokenPrices(c *gin.Context) {
	var req struct {
		Tokens []struct {
			Address   string `json:"address" binding:"required"`
			TokenName string `json:"token_name"`
		} `json:"tokens" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Tokens) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one token is required"})
		return
	}

	if len(req.Tokens) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "maximum 10 tokens allowed per request"})
		return
	}

	var results []interface{}
	var errors []string

	for _, token := range req.Tokens {
		// 验证地址格式
		if !strings.HasPrefix(token.Address, "0x") || len(token.Address) != 42 {
			errors = append(errors, fmt.Sprintf("invalid address format: %s", token.Address))
			continue
		}

		priceInfo, err := h.bscService.GetTokenPrice(token.Address, token.TokenName)
		if err != nil {
			logger.Warnf("Failed to get price for token %s: %v", token.Address, err)
			errors = append(errors, fmt.Sprintf("failed to get price for %s: %v", token.Address, err))
			continue
		}

		results = append(results, priceInfo)
	}

	response := gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["error_count"] = len(errors)
	}

	c.JSON(http.StatusOK, response)
}

// GetLiquidityInfo 获取流动性池信息
func (h *BSCHandler) GetLiquidityInfo(c *gin.Context) {
	tokenA := c.Param("tokenA")
	tokenB := c.Param("tokenB")

	if tokenA == "" || tokenB == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both tokenA and tokenB addresses are required"})
		return
	}

	// 验证地址格式
	if !strings.HasPrefix(tokenA, "0x") || len(tokenA) != 42 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tokenA address format"})
		return
	}

	if !strings.HasPrefix(tokenB, "0x") || len(tokenB) != 42 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tokenB address format"})
		return
	}

	// 获取流动性池地址
	liquidityPool, err := h.bscService.GetLiquidityPool(tokenA, tokenB)
	if err != nil {
		logger.Errorf("Failed to get liquidity pool: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取总流动性
	totalLiquidity, err := h.bscService.GetTotalLiquidity(tokenA, tokenB)
	if err != nil {
		logger.Errorf("Failed to get total liquidity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取代币信息
	tokenAInfo, err := h.bscService.GetTokenInfo(tokenA)
	if err != nil {
		logger.Warnf("Failed to get tokenA info: %v", err)
		tokenAInfo = &services.TokenInfo{Address: tokenA, Name: "Unknown", Symbol: "Unknown"}
	}

	tokenBInfo, err := h.bscService.GetTokenInfo(tokenB)
	if err != nil {
		logger.Warnf("Failed to get tokenB info: %v", err)
		tokenBInfo = &services.TokenInfo{Address: tokenB, Name: "Unknown", Symbol: "Unknown"}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"liquidity_pool":  liquidityPool,
			"total_liquidity": totalLiquidity,
			"token_a":         tokenAInfo,
			"token_b":         tokenBInfo,
			"pair_name":       fmt.Sprintf("%s/%s", tokenAInfo.Symbol, tokenBInfo.Symbol),
		},
	})
}
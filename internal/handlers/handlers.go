package handlers

import (
	"net/http"

	"chain/internal/config"
	"chain/internal/services"
	"chain/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ChainHandler 链上交互处理器
type ChainHandler struct {
	chainService *services.ChainService
}

// NewChainHandler 创建新的链上交互处理器
func NewChainHandler(cfg *config.Config) *ChainHandler {
	chainService := services.NewChainService(cfg)
	return &ChainHandler{
		chainService: chainService,
	}
}

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	chainHandler := NewChainHandler(cfg)

	// 健康检查
	router.GET("/health", healthCheck)

	// API路由组
	api := router.Group("/api/v1")
	{
		// 链上交互相关路由
		chain := api.Group("/chain")
		{
			chain.GET("/balance/:address", chainHandler.GetBalance)
			chain.POST("/transfer", chainHandler.Transfer)
			chain.GET("/transaction/:hash", chainHandler.GetTransaction)
			chain.POST("/contract/call", chainHandler.CallContract)
			chain.POST("/contract/deploy", chainHandler.DeployContract)
		}
	}
}

// healthCheck 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"message": "Chain service is running",
	})
}

// GetBalance 获取地址余额
func (h *ChainHandler) GetBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
		return
	}

	balance, err := h.chainService.GetBalance(address)
	if err != nil {
		logger.Errorf("Failed to get balance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": balance,
	})
}

// Transfer 转账
func (h *ChainHandler) Transfer(c *gin.Context) {
	var req struct {
		To     string `json:"to" binding:"required"`
		Amount string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txHash, err := h.chainService.Transfer(req.To, req.Amount)
	if err != nil {
		logger.Errorf("Failed to transfer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transaction_hash": txHash,
		"to": req.To,
		"amount": req.Amount,
	})
}

// GetTransaction 获取交易信息
func (h *ChainHandler) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction hash is required"})
		return
	}

	tx, err := h.chainService.GetTransaction(hash)
	if err != nil {
		logger.Errorf("Failed to get transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// CallContract 调用智能合约
func (h *ChainHandler) CallContract(c *gin.Context) {
	var req struct {
		ContractAddress string        `json:"contract_address" binding:"required"`
		MethodName      string        `json:"method_name" binding:"required"`
		Params          []interface{} `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.chainService.CallContract(req.ContractAddress, req.MethodName, req.Params)
	if err != nil {
		logger.Errorf("Failed to call contract: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// DeployContract 部署智能合约
func (h *ChainHandler) DeployContract(c *gin.Context) {
	var req struct {
		Bytecode string        `json:"bytecode" binding:"required"`
		ABI      string        `json:"abi" binding:"required"`
		Params   []interface{} `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contractAddress, txHash, err := h.chainService.DeployContract(req.Bytecode, req.ABI, req.Params)
	if err != nil {
		logger.Errorf("Failed to deploy contract: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contract_address": contractAddress,
		"transaction_hash": txHash,
	})
}
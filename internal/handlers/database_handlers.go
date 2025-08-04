package handlers

import (
	"net/http"
	"strconv"

	"chain/internal/database"
	"chain/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DatabaseHandler 数据库处理器
type DatabaseHandler struct {
	dbService *services.DatabaseService
}

// NewDatabaseHandler 创建数据库处理器
func NewDatabaseHandler(db *database.Database) *DatabaseHandler {
	return &DatabaseHandler{
		dbService: services.NewDatabaseService(db),
	}
}

// GetTransactionByHash 根据交易哈希获取交易
// @Summary 根据交易哈希获取交易
// @Description 通过交易哈希查询交易详情
// @Tags 交易
// @Accept json
// @Produce json
// @Param hash path string true "交易哈希"
// @Success 200 {object} models.Transaction
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/transactions/{hash} [get]
func (h *DatabaseHandler) GetTransactionByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "交易哈希不能为空"})
		return
	}

	tx, err := h.dbService.GetTransactionByHash(hash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "交易未找到"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, tx)
}

// GetTransactionsByAddress 获取地址相关的交易
// @Summary 获取地址相关的交易
// @Description 查询指定地址的所有交易记录
// @Tags 交易
// @Accept json
// @Produce json
// @Param address path string true "地址"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} models.Transaction
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/addresses/{address}/transactions [get]
func (h *DatabaseHandler) GetTransactionsByAddress(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址不能为空"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	txs, err := h.dbService.GetTransactionsByAddress(address, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": txs,
		"count":        len(txs),
		"limit":        limit,
		"offset":       offset,
	})
}

// GetBlockByNumber 根据区块号获取区块
// @Summary 根据区块号获取区块
// @Description 通过区块号查询区块详情
// @Tags 区块
// @Accept json
// @Produce json
// @Param number path int true "区块号"
// @Success 200 {object} models.Block
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/blocks/{number} [get]
func (h *DatabaseHandler) GetBlockByNumber(c *gin.Context) {
	numberStr := c.Param("number")
	number, err := strconv.ParseUint(numberStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的区块号"})
		return
	}

	block, err := h.dbService.GetBlockByNumber(number)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "区块未找到"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, block)
}

// GetBlockByHash 根据区块哈希获取区块
// @Summary 根据区块哈希获取区块
// @Description 通过区块哈希查询区块详情
// @Tags 区块
// @Accept json
// @Produce json
// @Param hash path string true "区块哈希"
// @Success 200 {object} models.Block
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/blocks/hash/{hash} [get]
func (h *DatabaseHandler) GetBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "区块哈希不能为空"})
		return
	}

	block, err := h.dbService.GetBlockByHash(hash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "区块未找到"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, block)
}

// GetLatestBlocks 获取最新的区块列表
// @Summary 获取最新的区块列表
// @Description 查询最新的区块列表
// @Tags 区块
// @Accept json
// @Produce json
// @Param limit query int false "限制数量" default(10)
// @Success 200 {array} models.Block
// @Failure 500 {object} map[string]string
// @Router /api/v1/blocks/latest [get]
func (h *DatabaseHandler) GetLatestBlocks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	blocks, err := h.dbService.GetLatestBlocks(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blocks": blocks,
		"count":  len(blocks),
		"limit":  limit,
	})
}

// GetAccountByAddress 根据地址获取账户信息
// @Summary 根据地址获取账户信息
// @Description 通过地址查询账户详情
// @Tags 账户
// @Accept json
// @Produce json
// @Param address path string true "账户地址"
// @Success 200 {object} models.Account
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/accounts/{address} [get]
func (h *DatabaseHandler) GetAccountByAddress(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址不能为空"})
		return
	}

	account, err := h.dbService.GetAccountByAddress(address)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "账户未找到"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetTokenByAddress 根据合约地址获取代币信息
// @Summary 根据合约地址获取代币信息
// @Description 通过合约地址查询代币详情
// @Tags 代币
// @Accept json
// @Produce json
// @Param address path string true "代币合约地址"
// @Success 200 {object} models.Token
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tokens/{address} [get]
func (h *DatabaseHandler) GetTokenByAddress(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "代币地址不能为空"})
		return
	}

	token, err := h.dbService.GetTokenByAddress(address)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "代币未找到"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, token)
}

// GetTokenBalancesByAccount 获取账户的所有代币余额
// @Summary 获取账户的所有代币余额
// @Description 查询指定账户的所有代币余额
// @Tags 代币余额
// @Accept json
// @Produce json
// @Param address path string true "账户地址"
// @Param chain_id query int false "链ID" default(56)
// @Success 200 {array} models.TokenBalance
// @Failure 500 {object} map[string]string
// @Router /api/v1/accounts/{address}/token-balances [get]
func (h *DatabaseHandler) GetTokenBalancesByAccount(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "地址不能为空"})
		return
	}

	chainIDStr := c.DefaultQuery("chain_id", "56")
	chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的链ID"})
		return
	}

	balances, err := h.dbService.GetTokenBalancesByAccount(address, chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balances": balances,
		"count":    len(balances),
		"address":  address,
		"chain_id": chainID,
	})
}

// GetStatistics 获取统计信息
// @Summary 获取统计信息
// @Description 获取区块链数据统计信息
// @Tags 统计
// @Accept json
// @Produce json
// @Param chain_id query int false "链ID" default(56)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/statistics [get]
func (h *DatabaseHandler) GetStatistics(c *gin.Context) {
	chainIDStr := c.DefaultQuery("chain_id", "56")
	chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的链ID"})
		return
	}

	txCount, err := h.dbService.GetTransactionCount(chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易数量失败: " + err.Error()})
		return
	}

	blockCount, err := h.dbService.GetBlockCount(chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块数量失败: " + err.Error()})
		return
	}

	accountCount, err := h.dbService.GetAccountCount(chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取账户数量失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chain_id":        chainID,
		"transaction_count": txCount,
		"block_count":      blockCount,
		"account_count":    accountCount,
	})
}

// SearchTransactions 搜索交易
// @Summary 搜索交易
// @Description 根据多种条件搜索交易
// @Tags 交易
// @Accept json
// @Produce json
// @Param hash query string false "交易哈希（支持模糊搜索）"
// @Param from query string false "发送方地址"
// @Param to query string false "接收方地址"
// @Param block_number query int false "区块号"
// @Param chain_id query int false "链ID" default(56)
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} models.Transaction
// @Failure 500 {object} map[string]string
// @Router /api/v1/transactions/search [get]
func (h *DatabaseHandler) SearchTransactions(c *gin.Context) {
	params := make(map[string]interface{})
	
	if hash := c.Query("hash"); hash != "" {
		params["hash"] = hash
	}
	if from := c.Query("from"); from != "" {
		params["from"] = from
	}
	if to := c.Query("to"); to != "" {
		params["to"] = to
	}
	if blockNumberStr := c.Query("block_number"); blockNumberStr != "" {
		if blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64); err == nil {
			params["block_number"] = blockNumber
		}
	}
	
	chainIDStr := c.DefaultQuery("chain_id", "56")
	if chainID, err := strconv.ParseUint(chainIDStr, 10, 64); err == nil {
		params["chain_id"] = chainID
	}
	
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	txs, err := h.dbService.SearchTransactions(params, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"transactions": txs,
		"count":        len(txs),
		"limit":        limit,
		"offset":       offset,
		"params":       params,
	})
}
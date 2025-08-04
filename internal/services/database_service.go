package services

import (
	"fmt"

	"chain/internal/database"
	"chain/internal/models"
	"gorm.io/gorm"
)

// DatabaseService 数据库服务
type DatabaseService struct {
	db *gorm.DB
}

// NewDatabaseService 创建数据库服务实例
func NewDatabaseService(db *database.Database) *DatabaseService {
	return &DatabaseService{
		db: db.GetDB(),
	}
}

// TransactionService 交易相关查询

// GetTransactionByHash 根据交易哈希获取交易
func (s *DatabaseService) GetTransactionByHash(hash string) (*models.Transaction, error) {
	var tx models.Transaction
	err := s.db.Where("hash = ?", hash).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// GetTransactionsByAddress 获取地址相关的交易
func (s *DatabaseService) GetTransactionsByAddress(address string, limit, offset int) ([]models.Transaction, error) {
	var txs []models.Transaction
	err := s.db.Where("from_address = ? OR to_address = ?", address, address).
		Order("block_number DESC").
		Limit(limit).
		Offset(offset).
		Find(&txs).Error
	return txs, err
}

// GetTransactionsByBlockNumber 获取指定区块的所有交易
func (s *DatabaseService) GetTransactionsByBlockNumber(blockNumber uint64) ([]models.Transaction, error) {
	var txs []models.Transaction
	err := s.db.Where("block_number = ?", blockNumber).Find(&txs).Error
	return txs, err
}

// CreateTransaction 创建交易记录
func (s *DatabaseService) CreateTransaction(tx *models.Transaction) error {
	return s.db.Create(tx).Error
}

// BlockService 区块相关查询

// GetBlockByNumber 根据区块号获取区块
func (s *DatabaseService) GetBlockByNumber(number uint64) (*models.Block, error) {
	var block models.Block
	err := s.db.Where("number = ?", number).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetBlockByHash 根据区块哈希获取区块
func (s *DatabaseService) GetBlockByHash(hash string) (*models.Block, error) {
	var block models.Block
	err := s.db.Where("hash = ?", hash).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetLatestBlocks 获取最新的区块列表
func (s *DatabaseService) GetLatestBlocks(limit int) ([]models.Block, error) {
	var blocks []models.Block
	err := s.db.Order("number DESC").Limit(limit).Find(&blocks).Error
	return blocks, err
}

// CreateBlock 创建区块记录
func (s *DatabaseService) CreateBlock(block *models.Block) error {
	return s.db.Create(block).Error
}

// AccountService 账户相关查询

// GetAccountByAddress 根据地址获取账户信息
func (s *DatabaseService) GetAccountByAddress(address string) (*models.Account, error) {
	var account models.Account
	err := s.db.Where("address = ?", address).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// CreateOrUpdateAccount 创建或更新账户信息
func (s *DatabaseService) CreateOrUpdateAccount(account *models.Account) error {
	var existingAccount models.Account
	err := s.db.Where("address = ? AND chain_id = ?", account.Address, account.ChainID).First(&existingAccount).Error
	
	if err == gorm.ErrRecordNotFound {
		// 账户不存在，创建新账户
		return s.db.Create(account).Error
	} else if err != nil {
		return err
	}
	
	// 账户存在，更新信息
	existingAccount.Balance = account.Balance
	existingAccount.Nonce = account.Nonce
	return s.db.Save(&existingAccount).Error
}

// TokenService 代币相关查询

// GetTokenByAddress 根据合约地址获取代币信息
func (s *DatabaseService) GetTokenByAddress(address string) (*models.Token, error) {
	var token models.Token
	err := s.db.Where("address = ?", address).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetTokensByChainID 获取指定链的所有代币
func (s *DatabaseService) GetTokensByChainID(chainID uint64, limit, offset int) ([]models.Token, error) {
	var tokens []models.Token
	err := s.db.Where("chain_id = ?", chainID).
		Limit(limit).
		Offset(offset).
		Find(&tokens).Error
	return tokens, err
}

// CreateToken 创建代币记录
func (s *DatabaseService) CreateToken(token *models.Token) error {
	return s.db.Create(token).Error
}

// TokenBalanceService 代币余额相关查询

// GetTokenBalancesByAccount 获取账户的所有代币余额
func (s *DatabaseService) GetTokenBalancesByAccount(accountAddress string, chainID uint64) ([]models.TokenBalance, error) {
	var balances []models.TokenBalance
	err := s.db.Joins("JOIN accounts ON accounts.id = token_balances.account_id").
		Joins("JOIN tokens ON tokens.id = token_balances.token_id").
		Where("accounts.address = ? AND token_balances.chain_id = ?", accountAddress, chainID).
		Preload("Account").
		Preload("Token").
		Find(&balances).Error
	return balances, err
}

// GetTokenBalance 获取特定账户的特定代币余额
func (s *DatabaseService) GetTokenBalance(accountAddress, tokenAddress string, chainID uint64) (*models.TokenBalance, error) {
	var balance models.TokenBalance
	err := s.db.Joins("JOIN accounts ON accounts.id = token_balances.account_id").
		Joins("JOIN tokens ON tokens.id = token_balances.token_id").
		Where("accounts.address = ? AND tokens.address = ? AND token_balances.chain_id = ?", 
			accountAddress, tokenAddress, chainID).
		Preload("Account").
		Preload("Token").
		First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// CreateOrUpdateTokenBalance 创建或更新代币余额
func (s *DatabaseService) CreateOrUpdateTokenBalance(balance *models.TokenBalance) error {
	var existingBalance models.TokenBalance
	err := s.db.Where("account_id = ? AND token_id = ? AND chain_id = ?", 
		balance.AccountID, balance.TokenID, balance.ChainID).First(&existingBalance).Error
	
	if err == gorm.ErrRecordNotFound {
		// 余额记录不存在，创建新记录
		return s.db.Create(balance).Error
	} else if err != nil {
		return err
	}
	
	// 余额记录存在，更新余额
	existingBalance.Balance = balance.Balance
	return s.db.Save(&existingBalance).Error
}

// StatisticsService 统计相关查询

// GetTransactionCount 获取交易总数
func (s *DatabaseService) GetTransactionCount(chainID uint64) (int64, error) {
	var count int64
	err := s.db.Model(&models.Transaction{}).Where("chain_id = ?", chainID).Count(&count).Error
	return count, err
}

// GetBlockCount 获取区块总数
func (s *DatabaseService) GetBlockCount(chainID uint64) (int64, error) {
	var count int64
	err := s.db.Model(&models.Block{}).Where("chain_id = ?", chainID).Count(&count).Error
	return count, err
}

// GetAccountCount 获取账户总数
func (s *DatabaseService) GetAccountCount(chainID uint64) (int64, error) {
	var count int64
	err := s.db.Model(&models.Account{}).Where("chain_id = ?", chainID).Count(&count).Error
	return count, err
}

// SearchTransactions 搜索交易（支持多种条件）
func (s *DatabaseService) SearchTransactions(params map[string]interface{}, limit, offset int) ([]models.Transaction, error) {
	query := s.db.Model(&models.Transaction{})
	
	if hash, ok := params["hash"]; ok {
		query = query.Where("hash LIKE ?", fmt.Sprintf("%%%s%%", hash))
	}
	
	if from, ok := params["from"]; ok {
		query = query.Where("from_address = ?", from)
	}
	
	if to, ok := params["to"]; ok {
		query = query.Where("to_address = ?", to)
	}
	
	if blockNumber, ok := params["block_number"]; ok {
		query = query.Where("block_number = ?", blockNumber)
	}
	
	if chainID, ok := params["chain_id"]; ok {
		query = query.Where("chain_id = ?", chainID)
	}
	
	var txs []models.Transaction
	err := query.Order("block_number DESC").Limit(limit).Offset(offset).Find(&txs).Error
	return txs, err
}
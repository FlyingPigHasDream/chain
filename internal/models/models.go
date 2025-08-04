package models

import (
	"time"

	"gorm.io/gorm"
)

// Transaction 交易记录模型
type Transaction struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Hash        string         `gorm:"uniqueIndex;size:66" json:"hash"`
	From        string         `gorm:"index;size:42" json:"from"`
	To          string         `gorm:"index;size:42" json:"to"`
	Value       string         `gorm:"type:varchar(78)" json:"value"` // 使用字符串存储大数值
	GasPrice    string         `gorm:"type:varchar(78)" json:"gas_price"`
	GasLimit    uint64         `json:"gas_limit"`
	GasUsed     uint64         `json:"gas_used"`
	Nonce       uint64         `json:"nonce"`
	BlockNumber uint64         `gorm:"index" json:"block_number"`
	BlockHash   string         `gorm:"index;size:66" json:"block_hash"`
	Status      uint           `gorm:"index" json:"status"` // 0: 失败, 1: 成功
	ChainID     uint64         `gorm:"index" json:"chain_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Block 区块信息模型
type Block struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Number       uint64         `gorm:"uniqueIndex" json:"number"`
	Hash         string         `gorm:"uniqueIndex;size:66" json:"hash"`
	ParentHash   string         `gorm:"index;size:66" json:"parent_hash"`
	Timestamp    uint64         `json:"timestamp"`
	GasLimit     uint64         `json:"gas_limit"`
	GasUsed      uint64         `json:"gas_used"`
	Miner        string         `gorm:"index;size:42" json:"miner"`
	Difficulty   string         `gorm:"type:varchar(78)" json:"difficulty"`
	TotalDifficulty string      `gorm:"type:varchar(78)" json:"total_difficulty"`
	Size         uint64         `json:"size"`
	TxCount      uint           `json:"tx_count"`
	ChainID      uint64         `gorm:"index" json:"chain_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Account 账户信息模型
type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Address   string         `gorm:"uniqueIndex;size:42" json:"address"`
	Balance   string         `gorm:"type:varchar(78)" json:"balance"`
	Nonce     uint64         `json:"nonce"`
	ChainID   uint64         `gorm:"index" json:"chain_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Token 代币信息模型
type Token struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Address     string         `gorm:"uniqueIndex;size:42" json:"address"`
	Name        string         `gorm:"size:100" json:"name"`
	Symbol      string         `gorm:"size:20" json:"symbol"`
	Decimals    uint8          `json:"decimals"`
	TotalSupply string         `gorm:"type:varchar(78)" json:"total_supply"`
	ChainID     uint64         `gorm:"index" json:"chain_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TokenBalance 代币余额模型
type TokenBalance struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	AccountID    uint           `gorm:"index" json:"account_id"`
	TokenID      uint           `gorm:"index" json:"token_id"`
	Balance      string         `gorm:"type:varchar(78)" json:"balance"`
	ChainID      uint64         `gorm:"index" json:"chain_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Account Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	Token   Token   `gorm:"foreignKey:TokenID" json:"token,omitempty"`
}

// TableName 设置表名
func (Transaction) TableName() string {
	return "transactions"
}

func (Block) TableName() string {
	return "blocks"
}

func (Account) TableName() string {
	return "accounts"
}

func (Token) TableName() string {
	return "tokens"
}

func (TokenBalance) TableName() string {
	return "token_balances"
}
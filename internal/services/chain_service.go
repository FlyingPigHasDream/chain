package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"chain/internal/config"
	"chain/pkg/logger"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ChainService 链上交互服务
type ChainService struct {
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
	chainID    *big.Int
	gasLimit   uint64
}

// NewChainService 创建新的链上交互服务
func NewChainService(cfg *config.Config) *ChainService {
	// 连接到以太坊节点
	client, err := ethclient.Dial(cfg.Chain.RPCURL)
	if err != nil {
		logger.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(cfg.Chain.PrivateKey, "0x"))
	if err != nil {
		logger.Fatalf("Failed to parse private key: %v", err)
	}

	// 获取公钥和地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Fatal("Failed to cast public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	chainID := big.NewInt(cfg.Chain.ChainID)

	logger.Infof("Chain service initialized with address: %s", address.Hex())

	return &ChainService{
		client:     client,
		privateKey: privateKey,
		publicKey:  publicKeyECDSA,
		address:    address,
		chainID:    chainID,
		gasLimit:   cfg.Chain.GasLimit,
	}
}

// GetBalance 获取地址余额
func (s *ChainService) GetBalance(address string) (string, error) {
	addr := common.HexToAddress(address)
	balance, err := s.client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get balance: %w", err)
	}

	// 转换为以太单位
	balanceInEther := new(big.Float)
	balanceInEther.SetString(balance.String())
	balanceInEther = balanceInEther.Quo(balanceInEther, big.NewFloat(1e18))

	return balanceInEther.String(), nil
}

// Transfer 转账
func (s *ChainService) Transfer(to, amount string) (string, error) {
	toAddress := common.HexToAddress(to)
	
	// 解析金额
	amountWei, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		// 尝试解析为以太单位
		amountFloat, ok := new(big.Float).SetString(amount)
		if !ok {
			return "", fmt.Errorf("invalid amount format")
		}
		amountWei, _ = new(big.Int).SetString(new(big.Float).Mul(amountFloat, big.NewFloat(1e18)).String(), 10)
	}

	// 获取nonce
	nonce, err := s.client.PendingNonceAt(context.Background(), s.address)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// 获取gas价格
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// 创建交易
	tx := types.NewTransaction(nonce, toAddress, amountWei, s.gasLimit, gasPrice, nil)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainID), s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 发送交易
	err = s.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	logger.Infof("Transaction sent: %s", signedTx.Hash().Hex())
	return signedTx.Hash().Hex(), nil
}

// GetTransaction 获取交易信息
func (s *ChainService) GetTransaction(hash string) (map[string]interface{}, error) {
	txHash := common.HexToHash(hash)
	tx, isPending, err := s.client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	result := map[string]interface{}{
		"hash":      tx.Hash().Hex(),
		"to":        tx.To().Hex(),
		"value":     tx.Value().String(),
		"gas":       tx.Gas(),
		"gas_price": tx.GasPrice().String(),
		"nonce":     tx.Nonce(),
		"pending":   isPending,
	}

	// 如果交易已确认，获取收据
	if !isPending {
		receipt, err := s.client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			result["status"] = receipt.Status
			result["block_number"] = receipt.BlockNumber.String()
			result["gas_used"] = receipt.GasUsed
		}
	}

	return result, nil
}

// CallContract 调用智能合约
func (s *ChainService) CallContract(contractAddress, methodName string, params []interface{}) (interface{}, error) {
	// 这里需要根据具体的合约ABI来实现
	// 这是一个简化的示例
	addr := common.HexToAddress(contractAddress)
	
	// 创建调用数据（这里需要根据实际ABI编码）
	callData := []byte{} // 实际实现中需要根据ABI编码方法调用
	
	msg := ethereum.CallMsg{
		To:   &addr,
		Data: callData,
	}

	result, err := s.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	return fmt.Sprintf("0x%x", result), nil
}

// DeployContract 部署智能合约
func (s *ChainService) DeployContract(bytecode, abiJSON string, params []interface{}) (string, string, error) {
	// 解析ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return "", "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// 获取nonce
	nonce, err := s.client.PendingNonceAt(context.Background(), s.address)
	if err != nil {
		return "", "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// 获取gas价格
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// 创建交易选项
	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.chainID)
	if err != nil {
		return "", "", fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = s.gasLimit
	auth.GasPrice = gasPrice

	// 部署合约
	address, tx, _, err := bind.DeployContract(auth, parsedABI, common.FromHex(bytecode), s.client, params...)
	if err != nil {
		return "", "", fmt.Errorf("failed to deploy contract: %w", err)
	}

	logger.Infof("Contract deployed at: %s, tx: %s", address.Hex(), tx.Hash().Hex())
	return address.Hex(), tx.Hash().Hex(), nil
}
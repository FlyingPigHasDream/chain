package services

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"chain/internal/config"
	"chain/pkg/logger"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BSCService BSC链交互服务
type BSCService struct {
	client   *ethclient.Client
	chainID  *big.Int
	gasLimit uint64
}

// TokenInfo 代币信息
type TokenInfo struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

// PriceInfo 价格信息
type PriceInfo struct {
	TokenAddress    string `json:"token_address"`
	TokenName       string `json:"token_name"`
	TokenSymbol     string `json:"token_symbol"`
	PriceInBNB      string `json:"price_in_bnb"`
	PriceInUSD      string `json:"price_in_usd"`
	LiquidityPool   string `json:"liquidity_pool"`
	TotalLiquidity  string `json:"total_liquidity"`
	Volume24h       string `json:"volume_24h"`
	PriceChange24h  string `json:"price_change_24h"`
}

// PancakeSwap V2 Router 合约地址
const (
	PancakeSwapV2Router = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
	PancakeSwapV2Factory = "0xcA143Ce32Fe78f1f7019d7d551a6402fC5350c73"
	WBNBAddress         = "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c"
	USDTAddress         = "0x55d398326f99059fF775485246999027B3197955"
	BUSDAddress         = "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
)

// ERC20 ABI (简化版)
const erc20ABI = `[
	{
		"constant": true,
		"inputs": [],
		"name": "name",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [{"name": "", "type": "string"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [{"name": "", "type": "uint8"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [{"name": "_owner", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "balance", "type": "uint256"}],
		"type": "function"
	}
]`

// PancakeSwap Router ABI (简化版)
const pancakeRouterABI = `[
	{
		"constant": true,
		"inputs": [
			{"name": "amountIn", "type": "uint256"},
			{"name": "path", "type": "address[]"}
		],
		"name": "getAmountsOut",
		"outputs": [{"name": "amounts", "type": "uint256[]"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "WETH",
		"outputs": [{"name": "", "type": "address"}],
		"type": "function"
	}
]`

// PancakeSwap Factory ABI (简化版)
const pancakeFactoryABI = `[
	{
		"constant": true,
		"inputs": [
			{"name": "tokenA", "type": "address"},
			{"name": "tokenB", "type": "address"}
		],
		"name": "getPair",
		"outputs": [{"name": "pair", "type": "address"}],
		"type": "function"
	}
]`

// Pair ABI (简化版)
const pairABI = `[
	{
		"constant": true,
		"inputs": [],
		"name": "getReserves",
		"outputs": [
			{"name": "reserve0", "type": "uint112"},
			{"name": "reserve1", "type": "uint112"},
			{"name": "blockTimestampLast", "type": "uint32"}
		],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "token0",
		"outputs": [{"name": "", "type": "address"}],
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "token1",
		"outputs": [{"name": "", "type": "address"}],
		"type": "function"
	}
]`

// NewBSCService 创建新的BSC服务
func NewBSCService(cfg *config.Config) *BSCService {
	// 连接到BSC节点
	client, err := ethclient.Dial(cfg.Chain.RPCURL)
	if err != nil {
		logger.Fatalf("Failed to connect to BSC client: %v", err)
	}

	chainID := big.NewInt(cfg.Chain.ChainID)
	logger.Infof("BSC service initialized with chain ID: %d", cfg.Chain.ChainID)

	return &BSCService{
		client:   client,
		chainID:  chainID,
		gasLimit: cfg.Chain.GasLimit,
	}
}

// GetTokenInfo 获取代币信息
func (s *BSCService) GetTokenInfo(tokenAddress string) (*TokenInfo, error) {
	addr := common.HexToAddress(tokenAddress)
	
	// 解析ERC20 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ERC20 ABI: %w", err)
	}

	// 获取代币名称
	nameData, err := parsedABI.Pack("name")
	if err != nil {
		return nil, fmt.Errorf("failed to pack name call: %w", err)
	}
	nameResult, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &addr,
		Data: nameData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call name: %w", err)
	}
	nameOutput, err := parsedABI.Unpack("name", nameResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack name: %w", err)
	}

	// 获取代币符号
	symbolData, err := parsedABI.Pack("symbol")
	if err != nil {
		return nil, fmt.Errorf("failed to pack symbol call: %w", err)
	}
	symbolResult, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &addr,
		Data: symbolData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call symbol: %w", err)
	}
	symbolOutput, err := parsedABI.Unpack("symbol", symbolResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack symbol: %w", err)
	}

	// 获取代币精度
	decimalsData, err := parsedABI.Pack("decimals")
	if err != nil {
		return nil, fmt.Errorf("failed to pack decimals call: %w", err)
	}
	decimalsResult, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &addr,
		Data: decimalsData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call decimals: %w", err)
	}
	decimalsOutput, err := parsedABI.Unpack("decimals", decimalsResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack decimals: %w", err)
	}

	return &TokenInfo{
		Address:  tokenAddress,
		Name:     nameOutput[0].(string),
		Symbol:   symbolOutput[0].(string),
		Decimals: decimalsOutput[0].(uint8),
	}, nil
}

// GetTokenPrice 获取代币价格
func (s *BSCService) GetTokenPrice(tokenAddress, tokenName string) (*PriceInfo, error) {
	// 获取代币信息
	tokenInfo, err := s.GetTokenInfo(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}

	// 如果提供了代币名称，验证是否匹配
	if tokenName != "" && !strings.EqualFold(tokenInfo.Name, tokenName) && !strings.EqualFold(tokenInfo.Symbol, tokenName) {
		return nil, fmt.Errorf("token name/symbol mismatch: expected %s, got %s/%s", tokenName, tokenInfo.Name, tokenInfo.Symbol)
	}

	// 获取BNB价格
	priceInBNB, err := s.getTokenPriceInBNB(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get price in BNB: %w", err)
	}

	// 获取BNB/USDT价格来计算USD价格
	bnbPriceInUSD, err := s.getBNBPriceInUSD()
	if err != nil {
		logger.Warnf("Failed to get BNB price in USD: %v", err)
		bnbPriceInUSD = big.NewFloat(0)
	}

	// 计算USD价格
	priceInUSD := new(big.Float).Mul(priceInBNB, bnbPriceInUSD)

	// 获取流动性池地址
	liquidityPool, err := s.getLiquidityPool(tokenAddress, WBNBAddress)
	if err != nil {
		logger.Warnf("Failed to get liquidity pool: %v", err)
		liquidityPool = ""
	}

	// 获取流动性信息
	totalLiquidity, err := s.getTotalLiquidity(tokenAddress, WBNBAddress)
	if err != nil {
		logger.Warnf("Failed to get total liquidity: %v", err)
		totalLiquidity = "0"
	}

	return &PriceInfo{
		TokenAddress:    tokenAddress,
		TokenName:       tokenInfo.Name,
		TokenSymbol:     tokenInfo.Symbol,
		PriceInBNB:      priceInBNB.String(),
		PriceInUSD:      priceInUSD.String(),
		LiquidityPool:   liquidityPool,
		TotalLiquidity:  totalLiquidity,
		Volume24h:       "0", // 需要额外的API来获取24小时交易量
		PriceChange24h:  "0", // 需要额外的API来获取24小时价格变化
	}, nil
}

// getTokenPriceInBNB 获取代币相对于BNB的价格
func (s *BSCService) getTokenPriceInBNB(tokenAddress string) (*big.Float, error) {
	// 解析Router ABI
	parsedABI, err := abi.JSON(strings.NewReader(pancakeRouterABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse router ABI: %w", err)
	}

	// 准备路径：token -> WBNB
	path := []common.Address{
		common.HexToAddress(tokenAddress),
		common.HexToAddress(WBNBAddress),
	}

	// 1个代币单位（考虑精度）
	amountIn := big.NewInt(1)
	amountIn = amountIn.Exp(big.NewInt(10), big.NewInt(18), nil) // 假设18位精度

	// 调用getAmountsOut
	data, err := parsedABI.Pack("getAmountsOut", amountIn, path)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getAmountsOut: %w", err)
	}

	routerAddr := common.HexToAddress(PancakeSwapV2Router)
	result, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &routerAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getAmountsOut: %w", err)
	}

	output, err := parsedABI.Unpack("getAmountsOut", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getAmountsOut: %w", err)
	}

	amounts := output[0].([]*big.Int)
	if len(amounts) < 2 {
		return nil, fmt.Errorf("invalid amounts output")
	}

	// 计算价格：输出BNB数量 / 输入代币数量
	priceInBNB := new(big.Float).Quo(
		new(big.Float).SetInt(amounts[1]),
		new(big.Float).SetInt(amountIn),
	)

	return priceInBNB, nil
}

// getBNBPriceInUSD 获取BNB相对于USD的价格
func (s *BSCService) getBNBPriceInUSD() (*big.Float, error) {
	// 解析Router ABI
	parsedABI, err := abi.JSON(strings.NewReader(pancakeRouterABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse router ABI: %w", err)
	}

	// 准备路径：WBNB -> USDT
	path := []common.Address{
		common.HexToAddress(WBNBAddress),
		common.HexToAddress(USDTAddress),
	}

	// 1 BNB
	amountIn := big.NewInt(1)
	amountIn = amountIn.Exp(big.NewInt(10), big.NewInt(18), nil)

	// 调用getAmountsOut
	data, err := parsedABI.Pack("getAmountsOut", amountIn, path)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getAmountsOut: %w", err)
	}

	routerAddr := common.HexToAddress(PancakeSwapV2Router)
	result, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &routerAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getAmountsOut: %w", err)
	}

	output, err := parsedABI.Unpack("getAmountsOut", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getAmountsOut: %w", err)
	}

	amounts := output[0].([]*big.Int)
	if len(amounts) < 2 {
		return nil, fmt.Errorf("invalid amounts output")
	}

	// USDT有18位精度，转换为USD价格
	priceInUSD := new(big.Float).Quo(
		new(big.Float).SetInt(amounts[1]),
		new(big.Float).SetInt(amountIn),
	)

	return priceInUSD, nil
}

// getLiquidityPool 获取流动性池地址
func (s *BSCService) getLiquidityPool(tokenA, tokenB string) (string, error) {
	// 解析Factory ABI
	parsedABI, err := abi.JSON(strings.NewReader(pancakeFactoryABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse factory ABI: %w", err)
	}

	// 调用getPair
	data, err := parsedABI.Pack("getPair", common.HexToAddress(tokenA), common.HexToAddress(tokenB))
	if err != nil {
		return "", fmt.Errorf("failed to pack getPair: %w", err)
	}

	factoryAddr := common.HexToAddress(PancakeSwapV2Factory)
	result, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &factoryAddr,
		Data: data,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call getPair: %w", err)
	}

	output, err := parsedABI.Unpack("getPair", result)
	if err != nil {
		return "", fmt.Errorf("failed to unpack getPair: %w", err)
	}

	pairAddress := output[0].(common.Address)
	return pairAddress.Hex(), nil
}

// getTotalLiquidity 获取总流动性
func (s *BSCService) getTotalLiquidity(tokenA, tokenB string) (string, error) {
	// 获取流动性池地址
	pairAddress, err := s.getLiquidityPool(tokenA, tokenB)
	if err != nil {
		return "0", err
	}

	if pairAddress == "0x0000000000000000000000000000000000000000" {
		return "0", fmt.Errorf("no liquidity pool found")
	}

	// 解析Pair ABI
	parsedABI, err := abi.JSON(strings.NewReader(pairABI))
	if err != nil {
		return "0", fmt.Errorf("failed to parse pair ABI: %w", err)
	}

	// 调用getReserves
	data, err := parsedABI.Pack("getReserves")
	if err != nil {
		return "0", fmt.Errorf("failed to pack getReserves: %w", err)
	}

	pairAddr := common.HexToAddress(pairAddress)
	result, err := s.client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &pairAddr,
		Data: data,
	}, nil)
	if err != nil {
		return "0", fmt.Errorf("failed to call getReserves: %w", err)
	}

	output, err := parsedABI.Unpack("getReserves", result)
	if err != nil {
		return "0", fmt.Errorf("failed to unpack getReserves: %w", err)
	}

	reserve0 := output[0].(*big.Int)
	reserve1 := output[1].(*big.Int)

	// 简单返回两个储备量的和作为总流动性指标
	totalLiquidity := new(big.Int).Add(reserve0, reserve1)
	return totalLiquidity.String(), nil
}

// GetLiquidityPool 获取流动性池地址（公开方法）
func (s *BSCService) GetLiquidityPool(tokenA, tokenB string) (string, error) {
	return s.getLiquidityPool(tokenA, tokenB)
}

// GetTotalLiquidity 获取总流动性（公开方法）
func (s *BSCService) GetTotalLiquidity(tokenA, tokenB string) (string, error) {
	return s.getTotalLiquidity(tokenA, tokenB)
}

// FindTokenByName 通过名称查找代币
func (s *BSCService) FindTokenByName(tokenName string) ([]*TokenInfo, error) {
	// 这里需要维护一个常用代币的映射表
	// 在实际应用中，可以连接到代币列表API或维护本地数据库
	commonTokens := map[string]string{
		"WBNB":     WBNBAddress,
		"BNB":      WBNBAddress,
		"USDT":     USDTAddress,
		"BUSD":     BUSDAddress,
		"CAKE":     "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82",
		"SAFEMOON": "0x8076C74C5e3F5852037F31Ff0093Eeb8c8ADd8D3",
	}

	var results []*TokenInfo
	for name, address := range commonTokens {
		if strings.Contains(strings.ToLower(name), strings.ToLower(tokenName)) {
			tokenInfo, err := s.GetTokenInfo(address)
			if err != nil {
				logger.Warnf("Failed to get token info for %s: %v", address, err)
				continue
			}
			results = append(results, tokenInfo)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no tokens found with name: %s", tokenName)
	}

	return results, nil
}
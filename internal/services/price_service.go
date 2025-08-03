package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"chain/internal/config"
)

// PriceService 价格服务
type PriceService struct {
	config     *config.Config
	httpClient *http.Client
}

// CoinGeckoPriceResponse CoinGecko API响应
type CoinGeckoPriceResponse struct {
	ID                string             `json:"id"`
	Symbol            string             `json:"symbol"`
	Name              string             `json:"name"`
	CurrentPrice      float64            `json:"current_price"`
	MarketCap         float64            `json:"market_cap"`
	MarketCapRank     int                `json:"market_cap_rank"`
	TotalVolume       float64            `json:"total_volume"`
	High24h           float64            `json:"high_24h"`
	Low24h            float64            `json:"low_24h"`
	PriceChange24h    float64            `json:"price_change_24h"`
	PriceChangePercent24h float64        `json:"price_change_percentage_24h"`
	LastUpdated       string             `json:"last_updated"`
}

// CryptoPriceInfo 加密货币价格信息
type CryptoPriceInfo struct {
	Symbol            string    `json:"symbol"`
	Name              string    `json:"name"`
	CurrentPrice      float64   `json:"current_price"`
	MarketCap         float64   `json:"market_cap"`
	Volume24h         float64   `json:"volume_24h"`
	PriceChange24h    float64   `json:"price_change_24h"`
	PriceChangePercent24h float64 `json:"price_change_percent_24h"`
	LastUpdated       time.Time `json:"last_updated"`
}

// NewPriceService 创建价格服务
func NewPriceService(cfg *config.Config) *PriceService {
	return &PriceService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetCryptoPrice 获取加密货币价格
func (p *PriceService) GetCryptoPrice(ctx context.Context, symbol string) (*CryptoPriceInfo, error) {
	// 使用CoinGecko免费API
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=1&page=1", strings.ToLower(symbol))
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var prices []CoinGeckoPriceResponse
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no price data found for symbol: %s", symbol)
	}

	price := prices[0]
	lastUpdated, _ := time.Parse(time.RFC3339, price.LastUpdated)

	return &CryptoPriceInfo{
		Symbol:                price.Symbol,
		Name:                  price.Name,
		CurrentPrice:          price.CurrentPrice,
		MarketCap:             price.MarketCap,
		Volume24h:             price.TotalVolume,
		PriceChange24h:        price.PriceChange24h,
		PriceChangePercent24h: price.PriceChangePercent24h,
		LastUpdated:           lastUpdated,
	}, nil
}

// GetMultipleCryptoPrices 批量获取加密货币价格
func (p *PriceService) GetMultipleCryptoPrices(ctx context.Context, symbols []string) (map[string]*CryptoPriceInfo, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	// 将符号转换为小写并用逗号连接
	idsParam := strings.ToLower(strings.Join(symbols, ","))
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=%d&page=1", idsParam, len(symbols))
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var prices []CoinGeckoPriceResponse
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := make(map[string]*CryptoPriceInfo)
	for _, price := range prices {
		lastUpdated, _ := time.Parse(time.RFC3339, price.LastUpdated)
		result[price.Symbol] = &CryptoPriceInfo{
			Symbol:                price.Symbol,
			Name:                  price.Name,
			CurrentPrice:          price.CurrentPrice,
			MarketCap:             price.MarketCap,
			Volume24h:             price.TotalVolume,
			PriceChange24h:        price.PriceChange24h,
			PriceChangePercent24h: price.PriceChangePercent24h,
			LastUpdated:           lastUpdated,
		}
	}

	return result, nil
}

// GetTopCryptoPrices 获取市值排名前N的加密货币价格
func (p *PriceService) GetTopCryptoPrices(ctx context.Context, limit int) ([]*CryptoPriceInfo, error) {
	if limit <= 0 || limit > 250 {
		limit = 10 // 默认获取前10名
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=%d&page=1", limit)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var prices []CoinGeckoPriceResponse
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var result []*CryptoPriceInfo
	for _, price := range prices {
		lastUpdated, _ := time.Parse(time.RFC3339, price.LastUpdated)
		result = append(result, &CryptoPriceInfo{
			Symbol:                price.Symbol,
			Name:                  price.Name,
			CurrentPrice:          price.CurrentPrice,
			MarketCap:             price.MarketCap,
			Volume24h:             price.TotalVolume,
			PriceChange24h:        price.PriceChange24h,
			PriceChangePercent24h: price.PriceChangePercent24h,
			LastUpdated:           lastUpdated,
		})
	}

	return result, nil
}

// SearchCrypto 搜索加密货币
func (p *PriceService) SearchCrypto(ctx context.Context, query string) ([]*CryptoPriceInfo, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// 使用CoinGecko搜索API
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/search?query=%s", query)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search crypto: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var searchResult struct {
		Coins []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"coins"`
	}

	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	if len(searchResult.Coins) == 0 {
		return []*CryptoPriceInfo{}, nil
	}

	// 获取搜索结果的价格信息（最多前5个）
	var ids []string
	for i, coin := range searchResult.Coins {
		if i >= 5 { // 限制结果数量
			break
		}
		ids = append(ids, coin.ID)
	}

	pricesMap, err := p.GetMultipleCryptoPrices(ctx, ids)
	if err != nil {
		return nil, err
	}

	var result []*CryptoPriceInfo
	for _, price := range pricesMap {
		result = append(result, price)
	}

	return result, nil
}

// GetPriceHistory 获取价格历史（简化版本）
func (p *PriceService) GetPriceHistory(ctx context.Context, symbol string, days int) ([]float64, error) {
	if days <= 0 || days > 365 {
		days = 7 // 默认7天
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=%d", strings.ToLower(symbol), days)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("price history request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var historyData struct {
		Prices [][]float64 `json:"prices"`
	}

	if err := json.Unmarshal(body, &historyData); err != nil {
		return nil, fmt.Errorf("failed to parse history response: %w", err)
	}

	var prices []float64
	for _, priceData := range historyData.Prices {
		if len(priceData) >= 2 {
			prices = append(prices, priceData[1]) // priceData[0]是时间戳，priceData[1]是价格
		}
	}

	return prices, nil
}
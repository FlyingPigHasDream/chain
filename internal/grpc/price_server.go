package grpc

import (
	"context"

	pb "chain/chain/proto"
	"chain/internal/services"
)

// PriceServer 价格服务gRPC实现
type PriceServer struct {
	pb.UnimplementedPriceServiceServer
	priceService *services.PriceService
}

// NewPriceServer 创建价格服务gRPC服务器
func NewPriceServer(priceService *services.PriceService) *PriceServer {
	return &PriceServer{
		priceService: priceService,
	}
}

// GetCryptoPrice 获取单个加密货币价格
func (s *PriceServer) GetCryptoPrice(ctx context.Context, req *pb.GetCryptoPriceRequest) (*pb.GetCryptoPriceResponse, error) {
	price, err := s.priceService.GetCryptoPrice(ctx, req.Symbol)
	if err != nil {
		return &pb.GetCryptoPriceResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetCryptoPriceResponse{
		Success: true,
		Price: &pb.CryptoPriceInfo{
			Symbol:                 price.Symbol,
			Name:                   price.Name,
			CurrentPrice:           price.CurrentPrice,
			MarketCap:              price.MarketCap,
			Volume_24H:             price.Volume24h,
			PriceChange_24H:        price.PriceChange24h,
			PriceChangePercent_24H: price.PriceChangePercent24h,
			LastUpdated:           price.LastUpdated.Format("2006-01-02 15:04:05"),
		},
	}, nil
}

// GetMultipleCryptoPrices 获取多个加密货币价格
func (s *PriceServer) GetMultipleCryptoPrices(ctx context.Context, req *pb.GetMultipleCryptoPricesRequest) (*pb.GetMultipleCryptoPricesResponse, error) {
	prices, err := s.priceService.GetMultipleCryptoPrices(ctx, req.Symbols)
	if err != nil {
		return &pb.GetMultipleCryptoPricesResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	pricesMap := make(map[string]*pb.CryptoPriceInfo)
	for symbol, price := range prices {
		pricesMap[symbol] = &pb.CryptoPriceInfo{
			Symbol:                 price.Symbol,
			Name:                   price.Name,
			CurrentPrice:           price.CurrentPrice,
			MarketCap:              price.MarketCap,
			Volume_24H:             price.Volume24h,
			PriceChange_24H:        price.PriceChange24h,
			PriceChangePercent_24H: price.PriceChangePercent24h,
			LastUpdated:           price.LastUpdated.Format("2006-01-02 15:04:05"),
		}
	}

	return &pb.GetMultipleCryptoPricesResponse{
		Success: true,
		Prices:  pricesMap,
	}, nil
}

// GetTopCryptoPrices 获取市值排名前N的加密货币价格
func (s *PriceServer) GetTopCryptoPrices(ctx context.Context, req *pb.GetTopCryptoPricesRequest) (*pb.GetTopCryptoPricesResponse, error) {
	prices, err := s.priceService.GetTopCryptoPrices(ctx, int(req.Limit))
	if err != nil {
		return &pb.GetTopCryptoPricesResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	var priceList []*pb.CryptoPriceInfo
	for _, price := range prices {
		priceList = append(priceList, &pb.CryptoPriceInfo{
			Symbol:                 price.Symbol,
			Name:                   price.Name,
			CurrentPrice:           price.CurrentPrice,
			MarketCap:              price.MarketCap,
			Volume_24H:             price.Volume24h,
			PriceChange_24H:        price.PriceChange24h,
			PriceChangePercent_24H: price.PriceChangePercent24h,
			LastUpdated:           price.LastUpdated.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.GetTopCryptoPricesResponse{
		Success: true,
		Prices:  priceList,
	}, nil
}

// SearchCrypto 搜索加密货币
func (s *PriceServer) SearchCrypto(ctx context.Context, req *pb.SearchCryptoRequest) (*pb.SearchCryptoResponse, error) {
	results, err := s.priceService.SearchCrypto(ctx, req.Query)
	if err != nil {
		return &pb.SearchCryptoResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	var resultList []*pb.CryptoPriceInfo
	for _, result := range results {
		resultList = append(resultList, &pb.CryptoPriceInfo{
			Symbol:                 result.Symbol,
			Name:                   result.Name,
			CurrentPrice:           result.CurrentPrice,
			MarketCap:              result.MarketCap,
			Volume_24H:             result.Volume24h,
			PriceChange_24H:        result.PriceChange24h,
			PriceChangePercent_24H: result.PriceChangePercent24h,
			LastUpdated:           result.LastUpdated.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.SearchCryptoResponse{
		Success: true,
		Results: resultList,
	}, nil
}

// GetPriceHistory 获取价格历史
func (s *PriceServer) GetPriceHistory(ctx context.Context, req *pb.GetPriceHistoryRequest) (*pb.GetPriceHistoryResponse, error) {
	prices, err := s.priceService.GetPriceHistory(ctx, req.Symbol, int(req.Days))
	if err != nil {
		return &pb.GetPriceHistoryResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetPriceHistoryResponse{
		Success: true,
		Prices:  prices,
	}, nil
}
package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"chain/internal/config"
	"chain/internal/services"
	pb "chain/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server gRPC服务器
type Server struct {
	grpcServer   *grpc.Server
	chainService *services.ChainService
	bscService   *services.BSCService
	config       *config.Config
}

// NewServer 创建新的gRPC服务器
func NewServer(cfg *config.Config) *Server {
	chainService := services.NewChainService(cfg)
	bscService := services.NewBSCService(cfg)

	s := &Server{
		grpcServer:   grpc.NewServer(),
		chainService: chainService,
		bscService:   bscService,
		config:       cfg,
	}

	// 注册服务
	pb.RegisterChainServiceServer(s.grpcServer, &chainServiceServer{chainService: chainService})
	pb.RegisterBSCServiceServer(s.grpcServer, &bscServiceServer{bscService: bscService})
	pb.RegisterHealthServiceServer(s.grpcServer, &healthServiceServer{})

	// 启用反射（用于调试）
	reflection.Register(s.grpcServer)

	return s
}

// Start 启动gRPC服务器
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.config.Server.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Printf("gRPC server starting on port %s", s.config.Server.GRPCPort)
	return s.grpcServer.Serve(lis)
}

// Stop 停止gRPC服务器
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

// chainServiceServer 链服务实现
type chainServiceServer struct {
	pb.UnimplementedChainServiceServer
	chainService *services.ChainService
}

func (s *chainServiceServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	balance, err := s.chainService.GetBalance(req.Address)
	if err != nil {
		return &pb.GetBalanceResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetBalanceResponse{
		Balance: balance,
		Address: req.Address,
		Success: true,
	}, nil
}

func (s *chainServiceServer) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	txHash, err := s.chainService.Transfer(req.To, req.Amount)
	if err != nil {
		return &pb.TransferResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.TransferResponse{
		TransactionHash: txHash,
		Success:         true,
	}, nil
}

func (s *chainServiceServer) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	tx, err := s.chainService.GetTransaction(req.Hash)
	if err != nil {
		return &pb.GetTransactionResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 类型断言从map[string]interface{}中提取值
	hash, _ := tx["hash"].(string)
	from, _ := tx["from"].(string)
	to, _ := tx["to"].(string)
	value, _ := tx["value"].(string)
	gasUsed, _ := tx["gas_used"].(uint64)
	gasPrice, _ := tx["gas_price"].(string)
	blockNumber, _ := tx["block_number"].(uint64)
	status, _ := tx["status"].(string)

	return &pb.GetTransactionResponse{
		Hash:        hash,
		From:        from,
		To:          to,
		Value:       value,
		GasUsed:     gasUsed,
		GasPrice:    gasPrice,
		BlockNumber: blockNumber,
		Status:      status,
		Success:     true,
	}, nil
}

func (s *chainServiceServer) CallContract(ctx context.Context, req *pb.CallContractRequest) (*pb.CallContractResponse, error) {
	// 转换参数类型
	params := make([]interface{}, len(req.Params))
	for i, param := range req.Params {
		params[i] = param
	}

	result, err := s.chainService.CallContract(req.ContractAddress, req.Method, params)
	if err != nil {
		return &pb.CallContractResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 类型断言结果
	resultStr := fmt.Sprintf("%v", result)

	return &pb.CallContractResponse{
		Result:  resultStr,
		Success: true,
	}, nil
}

func (s *chainServiceServer) DeployContract(ctx context.Context, req *pb.DeployContractRequest) (*pb.DeployContractResponse, error) {
	// 简化实现，暂时返回未实现错误
	return &pb.DeployContractResponse{
		Success: false,
		Error:   "DeployContract not implemented in gRPC service yet",
	}, nil
}

// bscServiceServer BSC服务实现
type bscServiceServer struct {
	pb.UnimplementedBSCServiceServer
	bscService *services.BSCService
}

func (s *bscServiceServer) GetTokenInfo(ctx context.Context, req *pb.GetTokenInfoRequest) (*pb.GetTokenInfoResponse, error) {
	tokenInfo, err := s.bscService.GetTokenInfo(req.Address)
	if err != nil {
		return &pb.GetTokenInfoResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetTokenInfoResponse{
		Token: &pb.TokenInfo{
			Address:  tokenInfo.Address,
			Name:     tokenInfo.Name,
			Symbol:   tokenInfo.Symbol,
			Decimals: uint32(tokenInfo.Decimals),
		},
		Success: true,
	}, nil
}

func (s *bscServiceServer) SearchToken(ctx context.Context, req *pb.SearchTokenRequest) (*pb.SearchTokenResponse, error) {
	tokens, err := s.bscService.FindTokenByName(req.Name)
	if err != nil {
		return &pb.SearchTokenResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	var pbTokens []*pb.TokenInfo
	for _, token := range tokens {
		pbTokens = append(pbTokens, &pb.TokenInfo{
			Address:  token.Address,
			Name:     token.Name,
			Symbol:   token.Symbol,
			Decimals: uint32(token.Decimals),
		})
	}

	return &pb.SearchTokenResponse{
		Tokens:  pbTokens,
		Success: true,
	}, nil
}

func (s *bscServiceServer) GetTokenPrice(ctx context.Context, req *pb.GetTokenPriceRequest) (*pb.GetTokenPriceResponse, error) {
	price, err := s.bscService.GetTokenPrice(req.Address, req.TokenName)
	if err != nil {
		return &pb.GetTokenPriceResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.GetTokenPriceResponse{
		Price: &pb.TokenPrice{
			Address:  price.TokenAddress,
			Name:     price.TokenName,
			Symbol:   price.TokenSymbol,
			PriceUsd: price.PriceInUSD,
			PriceBnb: price.PriceInBNB,
		},
		Success: true,
	}, nil
}

func (s *bscServiceServer) GetMultipleTokenPrices(ctx context.Context, req *pb.GetMultipleTokenPricesRequest) (*pb.GetMultipleTokenPricesResponse, error) {
	// 简化实现，逐个获取价格
	var pbPrices []*pb.TokenPrice
	for _, token := range req.Tokens {
		price, err := s.bscService.GetTokenPrice(token.Address, token.Name)
		if err != nil {
			continue // 跳过错误的代币
		}
		pbPrices = append(pbPrices, &pb.TokenPrice{
			Address:  price.TokenAddress,
			Name:     price.TokenName,
			Symbol:   price.TokenSymbol,
			PriceUsd: price.PriceInUSD,
			PriceBnb: price.PriceInBNB,
		})
	}

	return &pb.GetMultipleTokenPricesResponse{
		Prices:  pbPrices,
		Success: true,
	}, nil
}

func (s *bscServiceServer) GetLiquidityPool(ctx context.Context, req *pb.GetLiquidityPoolRequest) (*pb.GetLiquidityPoolResponse, error) {
	pairAddress, err := s.bscService.GetLiquidityPool(req.Token0, req.Token1)
	if err != nil {
		return &pb.GetLiquidityPoolResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	totalLiquidity, err := s.bscService.GetTotalLiquidity(req.Token0, req.Token1)
	if err != nil {
		totalLiquidity = "0" // 如果获取失败，设为0
	}

	return &pb.GetLiquidityPoolResponse{
		Pool: &pb.LiquidityPool{
			PairAddress:    pairAddress,
			Token0:         req.Token0,
			Token1:         req.Token1,
			Reserve0:       "0", // 简化实现
			Reserve1:       "0", // 简化实现
			TotalLiquidity: totalLiquidity,
		},
		Success: true,
	}, nil
}

// healthServiceServer 健康检查服务实现
type healthServiceServer struct {
	pb.UnimplementedHealthServiceServer
}

func (s *healthServiceServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:  "OK",
		Message: "Service is healthy",
	}, nil
}
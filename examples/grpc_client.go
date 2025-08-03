package main

import (
	"context"
	"log"
	"time"

	pb "chain/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 连接到gRPC服务器
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 创建客户端
	healthClient := pb.NewHealthServiceClient(conn)
	chainClient := pb.NewChainServiceClient(conn)
	bscClient := pb.NewBSCServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 测试健康检查
	log.Println("Testing Health Check...")
	healthResp, err := healthClient.Check(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		log.Printf("Health check response: %s - %s", healthResp.Status, healthResp.Message)
	}

	// 测试获取余额
	log.Println("\nTesting Get Balance...")
	balanceResp, err := chainClient.GetBalance(ctx, &pb.GetBalanceRequest{
		Address: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b",
	})
	if err != nil {
		log.Printf("Get balance failed: %v", err)
	} else {
		if balanceResp.Success {
			log.Printf("Balance: %s for address %s", balanceResp.Balance, balanceResp.Address)
		} else {
			log.Printf("Get balance error: %s", balanceResp.Error)
		}
	}

	// 测试BSC代币信息
	log.Println("\nTesting BSC Token Info...")
	tokenResp, err := bscClient.GetTokenInfo(ctx, &pb.GetTokenInfoRequest{
		Address: "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c", // WBNB
	})
	if err != nil {
		log.Printf("Get token info failed: %v", err)
	} else {
		if tokenResp.Success {
			token := tokenResp.Token
			log.Printf("Token: %s (%s) - %s, Decimals: %d", token.Name, token.Symbol, token.Address, token.Decimals)
		} else {
			log.Printf("Get token info error: %s", tokenResp.Error)
		}
	}

	// 测试代币价格
	log.Println("\nTesting Token Price...")
	priceResp, err := bscClient.GetTokenPrice(ctx, &pb.GetTokenPriceRequest{
		Address:   "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c", // WBNB
		TokenName: "WBNB",
	})
	if err != nil {
		log.Printf("Get token price failed: %v", err)
	} else {
		if priceResp.Success {
			price := priceResp.Price
			log.Printf("Price: %s (%s) - USD: %s, BNB: %s", price.Name, price.Symbol, price.PriceUsd, price.PriceBnb)
		} else {
			log.Printf("Get token price error: %s", priceResp.Error)
		}
	}

	// 测试搜索代币
	log.Println("\nTesting Search Token...")
	searchResp, err := bscClient.SearchToken(ctx, &pb.SearchTokenRequest{
		Name: "CAKE",
	})
	if err != nil {
		log.Printf("Search token failed: %v", err)
	} else {
		if searchResp.Success {
			log.Printf("Found %d tokens:", len(searchResp.Tokens))
			for _, token := range searchResp.Tokens {
				log.Printf("  - %s (%s): %s", token.Name, token.Symbol, token.Address)
			}
		} else {
			log.Printf("Search token error: %s", searchResp.Error)
		}
	}

	// 测试流动性池
	log.Println("\nTesting Liquidity Pool...")
	liquidityResp, err := bscClient.GetLiquidityPool(ctx, &pb.GetLiquidityPoolRequest{
		Token0: "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c", // WBNB
		Token1: "0x55d398326f99059fF775485246999027B3197955", // USDT
	})
	if err != nil {
		log.Printf("Get liquidity pool failed: %v", err)
	} else {
		if liquidityResp.Success {
			pool := liquidityResp.Pool
			log.Printf("Liquidity Pool: %s\n  Token0: %s\n  Token1: %s\n  Total Liquidity: %s",
				pool.PairAddress, pool.Token0, pool.Token1, pool.TotalLiquidity)
		} else {
			log.Printf("Get liquidity pool error: %s", liquidityResp.Error)
		}
	}

	log.Println("\nAll tests completed!")
}
# Chain Service gRPC API

本项目现在支持通过 gRPC 提供微服务，包含链上操作和 BSC 代币信息查询功能。

## 服务架构

项目提供两种服务模式：
- **HTTP REST API**: 传统的 REST 接口 (端口 8080)
- **gRPC API**: 高性能的 gRPC 接口 (端口 9090)

## gRPC 服务

### 服务列表

1. **ChainService**: 区块链基础操作
   - 获取账户余额
   - 发送代币转账
   - 获取交易信息
   - 智能合约调用
   - 智能合约部署

2. **BSCService**: BSC 链专用功能
   - 获取代币信息
   - 搜索代币
   - 获取代币价格
   - 批量获取代币价格
   - 获取流动性池信息

3. **HealthService**: 健康检查
   - 服务状态检查

### 启动 gRPC 服务

```bash
# 构建 gRPC 服务
make build-grpc

# 运行 gRPC 服务
make run-grpc

# 或者直接运行
go run cmd/grpc_main.go
```

服务将在 `localhost:9090` 启动。

### 测试 gRPC 服务

```bash
# 运行测试客户端
make test-grpc-client

# 或者直接运行
go run examples/grpc_client.go
```

### Protocol Buffers 定义

gRPC 服务基于 `proto/chain_service.proto` 定义，包含以下主要消息类型：

#### ChainService 消息
- `GetBalanceRequest/Response`: 获取余额
- `TransferRequest/Response`: 代币转账
- `GetTransactionRequest/Response`: 获取交易信息
- `CallContractRequest/Response`: 合约调用
- `DeployContractRequest/Response`: 合约部署

#### BSCService 消息
- `GetTokenInfoRequest/Response`: 获取代币信息
- `SearchTokenRequest/Response`: 搜索代币
- `GetTokenPriceRequest/Response`: 获取代币价格
- `GetMultipleTokenPricesRequest/Response`: 批量获取价格
- `GetLiquidityPoolRequest/Response`: 获取流动性池

#### 数据结构
- `TokenInfo`: 代币基础信息（名称、符号、地址、精度）
- `PriceInfo`: 价格信息（USD 价格、BNB 价格）
- `LiquidityPool`: 流动性池信息

### 客户端集成示例

#### Go 客户端

```go
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
    // 连接到 gRPC 服务器
    conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()
    
    // 创建客户端
    chainClient := pb.NewChainServiceClient(conn)
    bscClient := pb.NewBSCServiceClient(conn)
    
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()
    
    // 获取余额
    balanceResp, err := chainClient.GetBalance(ctx, &pb.GetBalanceRequest{
        Address: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b",
    })
    if err != nil {
        log.Printf("Error: %v", err)
    } else if balanceResp.Success {
        log.Printf("Balance: %s", balanceResp.Balance)
    }
    
    // 获取代币信息
    tokenResp, err := bscClient.GetTokenInfo(ctx, &pb.GetTokenInfoRequest{
        Address: "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c", // WBNB
    })
    if err != nil {
        log.Printf("Error: %v", err)
    } else if tokenResp.Success {
        token := tokenResp.Token
        log.Printf("Token: %s (%s)", token.Name, token.Symbol)
    }
}
```

### 开发工具

#### 重新生成 Protocol Buffers 代码

```bash
# 安装必要工具
make install-tools

# 生成 protobuf 代码
make proto-gen
```

#### 使用 grpcurl 测试（可选）

```bash
# 安装 grpcurl
brew install grpcurl

# 列出服务
grpcurl -plaintext localhost:9090 list

# 健康检查
grpcurl -plaintext localhost:9090 chain.HealthService/Check

# 获取余额
grpcurl -plaintext -d '{"address": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b"}' localhost:9090 chain.ChainService/GetBalance
```

### 配置

gRPC 服务端口可以通过以下方式配置：

1. **环境变量**: `SERVER_GRPC_PORT=9090`
2. **配置文件**: `configs/config.yaml` 中的 `server.grpc_port`

### 性能优势

gRPC 相比 REST API 具有以下优势：
- **更高性能**: 基于 HTTP/2 和 Protocol Buffers
- **强类型**: 编译时类型检查
- **流式传输**: 支持双向流
- **多语言支持**: 自动生成多种语言的客户端代码
- **更小的传输体积**: 二进制序列化

### 注意事项

1. gRPC 服务和 HTTP 服务可以同时运行
2. 网络连接问题可能导致某些 BSC 功能暂时不可用
3. 建议在生产环境中配置适当的超时和重试机制
4. 可以通过反射功能动态发现服务接口
# Chain Service - 链上交互微服务

这是一个基于Go语言开发的区块链交互微服务，提供了与以太坊及其兼容链进行交互的RESTful API接口。

## 功能特性

- 🔗 支持以太坊及其兼容链
- 🌟 **BSC链专项支持**
- 💰 账户余额查询
- 💸 代币转账功能
- 📋 交易信息查询
- 📜 智能合约调用
- 🚀 智能合约部署
- 💎 **BSC代币价格查询**
- 🏊 **流动性池信息查询**
- 🔍 **代币搜索功能**
- 📊 **PancakeSwap集成**
- 🐳 Docker容器化支持
- 📊 结构化日志记录
- ⚙️ 灵活的配置管理

## 项目结构

```
chain/
├── cmd/                    # 应用入口
│   └── main.go
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   ├── handlers/          # HTTP处理器
│   ├── services/          # 业务逻辑
│   └── server/            # HTTP服务器
├── pkg/                   # 公共包
│   └── logger/           # 日志工具
├── configs/               # 配置文件
│   └── config.yaml
├── .env.example          # 环境变量示例
├── Dockerfile            # Docker构建文件
├── docker-compose.yml    # Docker Compose配置
├── Makefile             # 构建脚本
└── go.mod               # Go模块文件
```

## 快速开始

### 环境要求

- Go 1.21+
- Docker (可选)
- 以太坊节点访问权限（如Infura）

### 安装依赖

```bash
go mod download
```

### 配置环境

1. 复制环境变量模板：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，填入你的配置：
```bash
# 区块链配置
CHAIN_RPC_URL=https://mainnet.infura.io/v3/your-project-id
CHAIN_PRIVATE_KEY=your-private-key-here
CHAIN_ID=1
```

### 运行服务

#### 方式1：直接运行
```bash
make run
# 或者
go run cmd/main.go
```

#### 方式2：使用Docker
```bash
make docker-build
make docker-run
```

#### 方式3：使用Docker Compose
```bash
make compose-up
```

## API接口

### 健康检查
```bash
GET /health
```

### 基础链功能

#### 获取账户余额
```bash
GET /api/v1/chain/balance/{address}
```

#### 代币转账
```bash
POST /api/v1/chain/transfer
{
  "to": "0x...",
  "amount": "1000000000000000000"
}
```

#### 查询交易信息
```bash
GET /api/v1/chain/transaction/{hash}
```

#### 调用智能合约
```bash
POST /api/v1/chain/contract/call
{
  "contract_address": "0x...",
  "method": "balanceOf",
  "params": ["0x..."]
}
```

#### 部署智能合约
```bash
POST /api/v1/chain/contract/deploy
{
  "bytecode": "0x608060405234801561001057600080fd5b50...",
  "constructor_params": []
}
```

### BSC专项功能

#### 获取代币信息
```bash
GET /api/v1/bsc/token/info/{address}
```

#### 通过名称搜索代币
```bash
GET /api/v1/bsc/token/search/{name}
```

#### 获取代币价格（通过地址）
```bash
GET /api/v1/bsc/token/price/{address}
```

#### 获取代币价格（通过地址和名称）
```bash
POST /api/v1/bsc/token/price
{
  "address": "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c",
  "token_name": "WBNB"
}
```

#### 批量获取代币价格
```bash
POST /api/v1/bsc/tokens/prices
{
  "tokens": [
    {"address": "0xbb4CdB9CBd36B01bD1cBaeBF2De08d9173bc095c"},
    {"address": "0x55d398326f99059fF775485246999027B3197955"}
  ]
}
```

#### 获取流动性池信息
```bash
GET /api/v1/bsc/liquidity/{token0}/{token1}
```

## 开发指南

### 代码格式化
```bash
make fmt
```

### 代码检查
```bash
make vet
```

### 运行测试
```bash
make test
```

### 热重载开发
```bash
# 安装air工具
make install-tools

# 启动热重载
make dev
```

## 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| PORT | 服务端口 | 8080 |
| HOST | 服务主机 | 0.0.0.0 |
| LOG_LEVEL | 日志级别 | info |
| CHAIN_RPC_URL | 区块链RPC地址 | - |
| CHAIN_PRIVATE_KEY | 私钥 | - |
| CHAIN_ID | 链ID | 1 |
| GAS_LIMIT | Gas限制 | 21000 |

### 配置文件

可以通过 `configs/config.yaml` 文件进行配置，环境变量优先级更高。

## 安全注意事项

⚠️ **重要提醒**：

1. **私钥安全**：绝不要将私钥提交到代码仓库中
2. **环境隔离**：生产环境和测试环境使用不同的私钥
3. **权限控制**：建议在生产环境中添加API认证
4. **网络安全**：使用HTTPS和防火墙保护服务

## 部署

### Docker部署

```bash
# 构建镜像
docker build -t chain-service .

# 运行容器
docker run -d \
  --name chain-service \
  -p 8080:8080 \
  -e CHAIN_RPC_URL=your-rpc-url \
  -e CHAIN_PRIVATE_KEY=your-private-key \
  chain-service
```

### Kubernetes部署

可以基于提供的Docker镜像创建Kubernetes部署配置。

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

如果你觉得这个项目有用，请给它一个 ⭐️！

## 更新日志

### v1.0.0
- 初始版本发布
- 支持基本的链上交互功能
- Docker容器化支持# chain

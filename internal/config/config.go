package config

import (
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Chain    ChainConfig    `mapstructure:"chain"`
	Database DatabaseConfig `mapstructure:"database"`
	Registry RegistryConfig `mapstructure:"registry"`
	LogLevel string         `mapstructure:"log_level"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port     string `mapstructure:"port" json:"port"`
	GRPCPort string `mapstructure:"grpc_port" json:"grpc_port"`
	Host     string `mapstructure:"host" json:"host"`
}

// ChainConfig 区块链配置
type ChainConfig struct {
	RPCURL     string `mapstructure:"rpc_url"`
	PrivateKey string `mapstructure:"private_key"`
	ChainID    int64  `mapstructure:"chain_id"`
	GasLimit   uint64 `mapstructure:"gas_limit"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Type      string `mapstructure:"type" json:"type"`           // etcd, consul, memory
	Endpoints string `mapstructure:"endpoints" json:"endpoints"` // 注册中心地址，多个用逗号分隔
}

// Load 加载配置
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认值和环境变量
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic("Failed to unmarshal config: " + err.Error())
	}

	return &cfg
}

// setDefaults 设置默认配置值
func setDefaults() {
	viper.SetDefault("server.port", getEnv("PORT", "8080"))
	viper.SetDefault("server.grpc_port", getEnv("GRPC_PORT", "9090"))
	viper.SetDefault("server.host", getEnv("HOST", "0.0.0.0"))
	viper.SetDefault("log_level", getEnv("LOG_LEVEL", "info"))
	viper.SetDefault("chain.rpc_url", getEnv("CHAIN_RPC_URL", "https://mainnet.infura.io/v3/your-project-id"))
	viper.SetDefault("chain.chain_id", getEnvInt("CHAIN_ID", 1))
	viper.SetDefault("chain.gas_limit", getEnvUint64("GAS_LIMIT", 21000))
	viper.SetDefault("registry.type", getEnv("REGISTRY_TYPE", "etcd"))
	viper.SetDefault("registry.endpoints", getEnv("REGISTRY_ENDPOINTS", "localhost:2379"))
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整型环境变量
func getEnvInt(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvUint64 获取uint64环境变量
func getEnvUint64(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if uintValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			return uintValue
		}
	}
	return defaultValue
}
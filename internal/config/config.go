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
	LogLevel string         `mapstructure:"log_level"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
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
	viper.SetDefault("server.host", getEnv("HOST", "0.0.0.0"))
	viper.SetDefault("log_level", getEnv("LOG_LEVEL", "info"))
	viper.SetDefault("chain.rpc_url", getEnv("CHAIN_RPC_URL", "https://mainnet.infura.io/v3/your-project-id"))
	viper.SetDefault("chain.chain_id", getEnvInt("CHAIN_ID", 1))
	viper.SetDefault("chain.gas_limit", getEnvUint64("GAS_LIMIT", 21000))
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
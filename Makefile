.PHONY: build run test clean docker-build docker-run deps

# 应用名称
APP_NAME=chain-service

# 构建应用
build:
	go build -o bin/$(APP_NAME) cmd/main.go

# 运行应用
run:
	go run cmd/main.go

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -rf bin/

# 下载依赖
deps:
	go mod download
	go mod tidy

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
vet:
	go vet ./...

# 构建Docker镜像
docker-build:
	docker build -t $(APP_NAME) .

# 运行Docker容器
docker-run:
	docker run -p 8080:8080 $(APP_NAME)

# 使用docker-compose启动
compose-up:
	docker-compose up -d

# 停止docker-compose
compose-down:
	docker-compose down

# 查看日志
logs:
	docker-compose logs -f

# 开发模式（热重载）
dev:
	air

# 安装开发工具
install-tools:
	go install github.com/cosmtrek/air@latest

# 生成API文档
docs:
	swag init -g cmd/main.go
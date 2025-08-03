.PHONY: build run test clean docker-build docker-run deps

# 应用名称
APP_NAME=chain-service

# 构建应用
build:
	go build -o bin/chain-service cmd/main.go

# 构建gRPC服务
build-grpc:
	go build -o bin/chain-grpc-service cmd/grpc_server.go

# 运行HTTP服务
run:
	go run cmd/main.go

# 运行gRPC服务
run-grpc:
	go run cmd/grpc_server.go

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
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成protobuf代码
proto-gen:
	export PATH=$$PATH:$$(go env GOPATH)/bin && protoc --go_out=. --go-grpc_out=. proto/chain_service.proto

# 测试gRPC客户端
test-grpc-client:
	go run examples/grpc_client.go

# 生成API文档
docs:
	swag init -g cmd/main.go
# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建HTTP服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o http-server cmd/main.go

# 构建gRPC服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpc-server cmd/grpc_server.go

# 运行阶段
FROM alpine:latest

# 安装ca证书
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/http-server .
COPY --from=builder /app/grpc-server .
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080 9090

# 默认运行HTTP服务，可通过环境变量切换
CMD ["./http-server"]
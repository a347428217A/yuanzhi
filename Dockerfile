# 使用多阶段构建减小镜像大小
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 安装 swag 工具
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 复制源代码
COPY . .

# 生成 Swagger 文档 (关键修复!)
RUN swag init

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o admin-api .

# 创建最终镜像
FROM alpine:3.18

# 安装CA证书（用于HTTPS请求）
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件和 Swagger 文档
COPY --from=builder /app/admin-api .
COPY --from=builder /app/docs ./docs  # 关键！

# 启动应用
CMD ["./admin-api"]
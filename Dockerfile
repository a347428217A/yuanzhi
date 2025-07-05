# 使用多阶段构建减小镜像大小
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .
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

# 从构建阶段复制二进制文件
COPY --from=builder /app/admin-api .

# 暴露端口（微信云托管默认使用80端口）
EXPOSE 80

# 设置环境变量（微信云托管会注入PORT环境变量）
ENV PORT=80

# 启动应用
CMD ["./admin-api"]
# 使用多阶段构建减小镜像大小
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 1. 先安装必要的工具和依赖
RUN apk add --no-cache wget tar git

# 2. 安装 swag 工具（使用二进制版本避免 Go 依赖问题）
RUN wget -q https://github.com/swaggo/swag/releases/download/v1.16.3/swag_1.16.3_Linux_x86_64.tar.gz && \
    tar -xvzf swag_1.16.3_Linux_x86_64.tar.gz && \
    mv swag /usr/local/bin/ && \
    rm swag_1.16.3_Linux_x86_64.tar.gz

# 3. 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 4. 复制源代码
COPY . .

# 5. 生成 Swagger 文档
RUN swag init --parseDependency --output ./docs

# 6. 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o admin-api .

# 创建最终镜像
FROM alpine:3.18

# 安装CA证书（用于HTTPS请求）
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件和文档
COPY --from=builder /app/admin-api .
COPY --from=builder /app/docs ./docs  # 复制文档目录

# 重要：云托管使用 PORT 环境变量，不要固定设置端口
# 暴露端口（仅作为文档说明，云托管会忽略）
EXPOSE 80

# 重要：不要设置 PORT 环境变量，使用云托管注入的值
# 启动应用
CMD ["./admin-api"]
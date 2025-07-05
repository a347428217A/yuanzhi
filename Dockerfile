# 阶段一：构建应用
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /usr/src/app

# 1. 复制依赖定义文件
COPY go.mod go.sum ./

# 下载依赖（使用国内代理加速）
RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

# 2. 复制所有源代码和文件
COPY . .

# 3. 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /usr/local/bin/admin-api .

# 阶段二：创建生产镜像
FROM alpine:3.18

# 安装基础依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# 设置上海时区
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -h /app

# 设置工作目录
WORKDIR /app
RUN chown -R appuser:appgroup /app
USER appuser

# 4. 复制二进制文件
COPY --from=builder --chown=appuser:appgroup /usr/local/bin/admin-api .

# 5. 复制配置文件（关键）
COPY --from=builder --chown=appuser:appgroup /usr/src/app/config.yaml .

# 6. 复制所有必需目录（确保完整结构）
COPY --from=builder --chown=appuser:appgroup /usr/src/app/certs ./certs/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/docs ./docs/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/common ./common/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/config ./config/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/controllers ./controllers/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/database ./database/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/middlewares ./middlewares/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/models ./models/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/payment ./payment/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/pkg ./pkg/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/routes ./routes/
COPY --from=builder --chown=appuser:appgroup /usr/src/app/utils ./utils/

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
    CMD wget -q --spider http://localhost:${PORT}/health || exit 1

# 暴露端口
EXPOSE 80

# 设置环境变量
ENV PORT=80

# 启动应用
CMD ["/app/admin-api"]
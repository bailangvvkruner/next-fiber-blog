# 多阶段构建：一个容器，最小化，全部使用最新版本避免安全漏洞
# 目标：Go后端提供API服务，并直接提供前端静态文件

# 第一阶段：构建 Go 后端（使用最新Go版本）
FROM golang:alpine AS go-builder
WORKDIR /go/src/app
# 复制依赖文件以利用Docker缓存
COPY go-fiber-server/go.mod go-fiber-server/go.sum ./
# 更新到最新模块版本，修复潜在漏洞
RUN go mod download && go mod tidy
COPY go-fiber-server/ .
# 构建为静态二进制文件，减少运行时依赖
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o next-fiber-blog ./cmd/go-fiber-server

# 第二阶段：构建 Next.js 前台（使用最新Node LTS）
FROM node:lts-alpine AS next-builder
WORKDIR /app
# 只复制package.json，因为package-lock.json可能不存在
COPY blog-front/package.json ./
# 安装最新依赖，使用--legacy-peer-deps处理peer依赖问题
RUN npm install --legacy-peer-deps --no-audit
# 单独运行安全审计但不强制修复（仅显示报告）
RUN npm audit || true
COPY blog-front/ .
# Next.js需要构建为静态文件
RUN npm run build

# 第三阶段：构建 Vue 管理后台（使用最新Node LTS）
FROM node:lts-alpine AS vue-builder
WORKDIR /app
# 只复制package.json，因为package-lock.json可能不存在
COPY blog-admin/package.json ./
# 安装最新依赖，使用--legacy-peer-deps处理peer依赖问题
RUN npm install --legacy-peer-deps --no-audit
# 单独运行安全审计但不强制修复（仅显示报告）
RUN npm audit || true
COPY blog-admin/ .
RUN npm run build

# 第四阶段：运行阶段 - 使用最精简的Alpine镜像
FROM alpine:latest
WORKDIR /app

# 更新系统并安装必要的CA证书（用于HTTPS请求）
RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*

# 1. 复制 Go 后端二进制文件（使用项目名）
COPY --from=go-builder /go/src/app/next-fiber-blog /app/

# 2. 复制 Next.js 前台构建产物
# 注意：根据next.config.mjs配置，Next.js可能输出到不同目录
# 这里假设标准构建输出到.next目录
COPY --from=next-builder /app/.next /app/static/front/.next
COPY --from=next-builder /app/public /app/static/front/public
COPY --from=next-builder /app/package.json /app/static/front/

# 3. 复制 Vue 管理后台构建产物
COPY --from=vue-builder /app/dist /app/static/admin

# 4. 复制Go后端配置文件（如果存在）
COPY go-fiber-server/configs/ /configs/

# 暴露Go服务端口（默认4000）
EXPOSE 4000

# 健康检查：检查API端点
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:4000/api/open/site/configuration || exit 1

# 启动Go服务
WORKDIR /app
CMD ["./next-fiber-blog", "-conf", "/configs/config-prod.yaml"]

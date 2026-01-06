# 阶段1：编译Go代码（基于官方golang镜像）
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum，下载依赖（利用Docker缓存，避免重复下载）
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源码文件
COPY . .

# 编译Go项目，生成可执行文件（命名为go-web-mvc）
# CGO_ENABLED=0 关闭CGO，生成静态二进制文件；GOOS=linux 指定运行系统为Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o go-web-mvc main.go

# 阶段2：运行可执行文件（基于alpine镜像，体积更小）
FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 从阶段1复制编译好的可执行文件到当前镜像
COPY --from=builder /app/go-web-mvc .

# 暴露后端服务端口（go-web-mvc默认端口是8080，需和Nginx配置中的端口一致）
EXPOSE 8080

# 启动后端服务
CMD ["./go-web-mvc"]
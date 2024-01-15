# 使用官方的 Go 镜像作为构建环境
FROM golang:1.16 as builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.* ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o yujian .

# 使用 alpine 镜像作为基础镜像
FROM alpine:latest  

# 设置工作目录
WORKDIR /root/

# 从构建者镜像中复制编译好的应用
COPY --from=builder /app/yujian .

# 从构建者镜像中复制所需的 HTML 和其他资源文件
COPY --from=builder /app/userprofile.html .
COPY --from=builder /app/uploadpage.html .
# 如果有其他文件或目录也需要被复制，确保在这里添加它们

# 暴露端口
EXPOSE 80

# 运行应用
CMD ["./yujian"]

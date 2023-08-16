# syntax=docker/dockerfile:1

# 设定构建步骤所使用的来源镜像为基于 Alpine 发行版的 Go 1.19 版本镜像
FROM golang:1.21-alpine as builder

ARG VERSION

# 设定 Go 使用 模块化依赖 管理方式：GO111MODULE
RUN GO111MODULE=on

# 创建路径 /app
RUN mkdir /app

# 复制当前目录下 perobot 到 /app/perobot
COPY . /app/perobot

# 切换到 /app/perobot 目录
WORKDIR /app/perobot

RUN go env
RUN go env -w CGO_ENABLED=0
RUN go build -a -o "release/pero" "github.com/nekomeowww/perobot/cmd/pero"

# 设定运行步骤所使用的镜像为基于 Alpine 发行版镜像
FROM alpine as runner

# 创建路径 /app
RUN mkdir /app
# 创建路径 /app/perobot/bin
RUN mkdir -p /app/perobot/bin

COPY --from=builder /app/perobot/release/pero /app/perobot/bin/

# 入点是编译好的 hyphen 应用程序
ENTRYPOINT [ "/app/perobot/bin/pero" ]

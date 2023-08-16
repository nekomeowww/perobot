# syntax=docker/dockerfile:1

# --- builder ---
FROM golang:1.21 as builder

RUN GO111MODULE=on

RUN mkdir /app

COPY . /app/perobot

WORKDIR /app/perobot

RUN go env
RUN go env -w CGO_ENABLED=0
RUN go build -a -o "release/pero" "github.com/nekomeowww/perobot/cmd/pero"


FROM debian as runner

RUN apt update && apt upgrade -y && apt install -y ca-certificates curl && update-ca-certificates

COPY --from=builder /app/perobot/release/pero /app/perobot/bin/

# 入点是编译好的 hyphen 应用程序
ENTRYPOINT [ "/usr/local/bin/pero" ]

# perobot

A pero bot, it peros a lot, maybe some purrrr as well.

## Usage

### Run

```shell
TELEGRAM_BOT_TOKEN=<Telegram Bot API Token> PIXIV_PHPSESSID=<Pixiv Cookie> pero
```

### Run with Docker

```shell
docker run -it --rm -e TELEGRAM_BOT_TOKEN=<Telegram Bot API Token> -e PIXIV_PHPSESSID=<Pixiv Cookie> pero nekomeowww/perobot:latest
```

### Run with docker-compose

Remember to replace your token and cookie in `docker-compose.yml`

```shell
docker-compose up -d
```

## Build on your own

### Build with go

```shell
go build -a -o "release/pero" "github.com/nekomeowww/perobot/cmd/pero"
```

### Build with Docker

```shell
docker buildx build --platform linux/arm64,linux/amd64 -t <tag> -f Dockerfile .
```

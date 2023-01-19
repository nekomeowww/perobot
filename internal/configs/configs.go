package configs

import "os"

const (
	EnvTelegramBotToken = "TELEGRAM_BOT_TOKEN"
)

type Config struct {
	TelegramBotToken string
}

func NewConfig() func() *Config {
	return func() *Config {
		return &Config{
			TelegramBotToken: os.Getenv(EnvTelegramBotToken),
		}
	}
}

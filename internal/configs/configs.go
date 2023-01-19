package configs

import "os"

const (
	EnvTelegramBotToken = "TELEGRAM_BOT_TOKEN"
	EnvPixivPHPSESSID   = "PIXIV_PHPSESSID"
)

type Config struct {
	TelegramBotToken string
	PixivPHPSESSID   string
}

func NewConfig() func() *Config {
	return func() *Config {
		return &Config{
			TelegramBotToken: os.Getenv(EnvTelegramBotToken),
			PixivPHPSESSID:   os.Getenv(EnvPixivPHPSESSID),
		}
	}
}

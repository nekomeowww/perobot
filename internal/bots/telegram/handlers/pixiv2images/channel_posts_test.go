package pixiv2images

import (
	"log"
	"os"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nekomeowww/perobot/internal/configs"
	"github.com/nekomeowww/perobot/internal/lib"
	"github.com/nekomeowww/perobot/internal/thirdparty"
	"github.com/nekomeowww/perobot/pkg/handler"
	"github.com/stretchr/testify/assert"
)

var h *Handler

func TestMain(m *testing.M) {
	config := configs.NewConfig()()
	logger := lib.NewLogger()()
	pixivPublic, err := thirdparty.NewPixivPublic()(thirdparty.NewPixivPublicParam{
		Config: config,
		Logger: logger,
	})
	if err != nil {
		log.Fatal(err)
	}

	h = NewHandler()(NewHandlerParam{
		Logger: logger,
		Pixiv:  pixivPublic,
	})

	os.Exit(m.Run())
}

func TestIllustIDFromText(t *testing.T) {
	artworkID := IllustIDFromText("https://www.pixiv.net/artworks/1234")
	assert.Equal(t, "1234", artworkID)
}

func TestHandleChannelPostPixivToImages(t *testing.T) {
	h.HandleChannelPostPixivToImages(handler.NewContext(&tgbotapi.BotAPI{}, tgbotapi.Update{
		ChannelPost: &tgbotapi.Message{
			Text: "https://www.pixiv.net/artworks/1234",
			Chat: &tgbotapi.Chat{
				ID: 1234,
			},
		},
	}))
}

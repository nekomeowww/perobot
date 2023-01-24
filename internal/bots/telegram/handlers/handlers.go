package handlers

import (
	"github.com/nekomeowww/perobot/internal/bots/telegram/dispatcher"
	"github.com/nekomeowww/perobot/internal/bots/telegram/handlers/pixiv2images"
	"github.com/nekomeowww/perobot/internal/bots/telegram/handlers/tweet2images"
	"github.com/nekomeowww/perobot/pkg/handler"
	"go.uber.org/fx"
)

func NewModules() fx.Option {
	return fx.Options(
		fx.Provide(NewHandlers()),
		fx.Provide(tweet2images.NewHandler()),
		fx.Provide(pixiv2images.NewHandler()),
	)
}

type NewHandlersParam struct {
	fx.In

	Dispatcher          *dispatcher.Dispatcher
	Tweet2ImagesHandler *tweet2images.Handler
	Pixiv2ImagesHandler *pixiv2images.Handler
}

type Handlers struct {
	Dispatcher *dispatcher.Dispatcher

	MessageHandlers     []handler.HandleFunc
	ChannelPostHandlers []handler.HandleFunc
}

func NewHandlers() func(param NewHandlersParam) *Handlers {
	return func(param NewHandlersParam) *Handlers {
		return &Handlers{
			Dispatcher: param.Dispatcher,
			MessageHandlers: []handler.HandleFunc{
				param.Tweet2ImagesHandler.HandleMessageAutomaticForwardedFromLinkedChannel,
				param.Pixiv2ImagesHandler.HandleMessageAutomaticForwardedFromLinkedChannel,
			},
			ChannelPostHandlers: []handler.HandleFunc{
				param.Tweet2ImagesHandler.HandleChannelPostTweetToImages,
				param.Pixiv2ImagesHandler.HandleChannelPostPixivToImages,
			},
		}
	}
}

func (h *Handlers) RegisterHandlers() {
	for _, handler := range h.MessageHandlers {
		h.Dispatcher.RegisterOneMessageHandler(handler)
	}
	for _, handler := range h.ChannelPostHandlers {
		h.Dispatcher.RegisterOneChannelPostHandler(handler)
	}
}

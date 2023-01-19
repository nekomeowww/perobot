package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"

	"github.com/nekomeowww/perobot/internal/bots/telegram/dispatcher"
	"github.com/nekomeowww/perobot/internal/bots/telegram/handlers"
	"github.com/nekomeowww/perobot/internal/configs"
	"github.com/nekomeowww/perobot/pkg/handler"
	"github.com/nekomeowww/perobot/pkg/logger"
	"github.com/nekomeowww/perobot/pkg/utils"
)

func NewModules() fx.Option {
	return fx.Options(
		fx.Provide(NewBot()),
		fx.Options(dispatcher.NewModules()),
		fx.Options(handlers.NewModules()),
	)
}

type NewBotParam struct {
	fx.In

	Lifecycle fx.Lifecycle

	Config     *configs.Config
	Logger     *logger.Logger
	Dispatcher *dispatcher.Dispatcher
	Handlers   *handlers.Handlers
}

type Bot struct {
	*tgbotapi.BotAPI

	Config     *configs.Config
	Logger     *logger.Logger
	Dispatcher *dispatcher.Dispatcher

	alreadyClose bool
	closeChan    chan struct{}
}

func NewBot() func(param NewBotParam) (*Bot, error) {
	return func(param NewBotParam) (*Bot, error) {
		if param.Config.TelegramBotToken == "" {
			param.Logger.Fatal("must supply a valid telegram bot token in configs or environment variable")
		}

		b, err := tgbotapi.NewBotAPI(param.Config.TelegramBotToken)
		if err != nil {
			return nil, err
		}

		bot := &Bot{
			BotAPI:     b,
			Logger:     param.Logger,
			Dispatcher: param.Dispatcher,
			closeChan:  make(chan struct{}),
		}

		param.Lifecycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				bot.StopPull(ctx)
				return nil
			},
		})

		param.Logger.Infof("authorized on bot @%s", bot.Self.UserName)
		param.Handlers.RegisterHandlers()
		return bot, nil
	}
}

func (b *Bot) StopPull(ctx context.Context) {
	if b.alreadyClose {
		return
	}

	_ = utils.Invoke0(func() error {
		b.alreadyClose = true
		b.StopReceivingUpdates()
		b.closeChan <- struct{}{}
		close(b.closeChan)

		return nil
	}, utils.WithContext(ctx))
}

func (b *Bot) PullUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.GetUpdatesChan(u)
	for {
		if b.alreadyClose {
			b.Logger.Info("stopped to receiving updates")
			return
		}

		select {
		case update := <-updates:
			if update.Message != nil { // If we got a message
				b.Logger.Infof("Message [%s] %s", update.Message.From.UserName, update.Message.Text)
				b.Dispatcher.DispatchMessage(handler.NewContext(b.BotAPI, update))
			}
			if update.ChannelPost != nil {
				b.Logger.Infof("Channel Post [%s] %s", update.ChannelPost.Chat.Title, update.ChannelPost.Text)
				b.Dispatcher.DispatchChannelPost(handler.NewContext(b.BotAPI, update))
			}
		case <-b.closeChan:
			b.Logger.Info("stopped to receiving updates")
			return
		}
	}
}

func Run() func(bot *Bot) {
	return func(bot *Bot) {
		go bot.PullUpdates()
	}
}

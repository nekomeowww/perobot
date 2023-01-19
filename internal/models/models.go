package models

import (
	"github.com/nekomeowww/perobot/internal/models/twitter"
	"go.uber.org/fx"
)

func NewModules() fx.Option {
	return fx.Options(
		fx.Provide(twitter.NewModel()),
	)
}

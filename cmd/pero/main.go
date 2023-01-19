package main

import (
	"context"
	"log"
	"time"

	"github.com/nekomeowww/perobot/internal/bots/telegram"
	"github.com/nekomeowww/perobot/internal/configs"
	"github.com/nekomeowww/perobot/internal/lib"
	"github.com/nekomeowww/perobot/internal/models"
	"github.com/nekomeowww/perobot/internal/thirdparty"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(fx.Options(
		fx.Provide(configs.NewConfig()),
		fx.Options(lib.NewModules()),
		fx.Options(models.NewModules()),
		fx.Options(thirdparty.NewModules()),
		fx.Options(telegram.NewModules()),
		fx.Invoke(telegram.Run()),
	))

	app.Run()
	stopCtx, stopCtxCancel := context.WithTimeout(context.Background(), time.Second*15)
	defer stopCtxCancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}

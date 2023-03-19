package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"go.uber.org/fx"

	"github.com/nekomeowww/perobot/internal/bots/telegram"
	"github.com/nekomeowww/perobot/internal/configs"
	"github.com/nekomeowww/perobot/internal/lib"
	"github.com/nekomeowww/perobot/internal/models"
	"github.com/nekomeowww/perobot/internal/thirdparty"
)

func main() {
	app := fx.New(fx.Options(
		fx.Provide(configs.NewConfig()),
		fx.Options(lib.NewModules()),
		fx.Options(models.NewModules()),
		fx.Options(thirdparty.NewModules()),
		fx.Options(telegram.NewModules()),
		fx.Invoke(telegram.Run()),
		fx.Invoke(func() {
			err := http.ListenAndServe(":6060", nil)
			if err != nil {
				log.Println(err)
			}
		}),
	))

	app.Run()
	stopCtx, stopCtxCancel := context.WithTimeout(context.Background(), time.Second*15)
	defer stopCtxCancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}

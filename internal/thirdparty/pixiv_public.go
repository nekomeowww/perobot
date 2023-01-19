package thirdparty

import (
	"github.com/nekomeowww/perobot/internal/configs"
	"github.com/nekomeowww/perobot/pkg/logger"
	pixiv_public "github.com/nekomeowww/perobot/pkg/pixiv/public"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type NewPixivPublicParam struct {
	fx.In

	Logger *logger.Logger
	Config *configs.Config
}

type PixivPublic struct {
	*pixiv_public.Client
}

func NewPixivPublic() func(param NewPixivPublicParam) (*PixivPublic, error) {
	return func(param NewPixivPublicParam) (*PixivPublic, error) {
		client, err := pixiv_public.NewClient(
			param.Config.PixivPHPSESSID,
			pixiv_public.WithLogger(logrus.NewEntry(param.Logger.Logger)),
		)
		if err != nil {
			param.Logger.Fatal(err)
		}

		return &PixivPublic{
			Client: client,
		}, nil
	}
}

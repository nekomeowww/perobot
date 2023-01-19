package thirdparty

import (
	"github.com/nekomeowww/perobot/pkg/logger"
	twitter_public "github.com/nekomeowww/perobot/pkg/twitter/public"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type NewTwitterPublicParam struct {
	fx.In

	Logger *logger.Logger
}

type TwitterPublic struct {
	*twitter_public.Client
}

func NewTwitterPublic() func(param NewTwitterPublicParam) (*TwitterPublic, error) {
	return func(param NewTwitterPublicParam) (*TwitterPublic, error) {
		client, err := twitter_public.NewClient(
			twitter_public.WithLogger(logrus.NewEntry(param.Logger.Logger)),
		)
		if err != nil {
			return nil, err
		}

		return &TwitterPublic{
			Client: client,
		}, nil
	}
}

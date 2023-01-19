package thirdparty

import (
	twitter_public "github.com/nekomeowww/perobot/pkg/twitter/public"
)

type TwitterPublic struct {
	*twitter_public.Client
}

func NewTwitterPublic() func() (*TwitterPublic, error) {
	return func() (*TwitterPublic, error) {
		client, err := twitter_public.NewClient()
		if err != nil {
			return nil, err
		}

		return &TwitterPublic{
			Client: client,
		}, nil
	}
}

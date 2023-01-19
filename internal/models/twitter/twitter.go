package twitter

import (
	"github.com/nekomeowww/perobot/internal/thirdparty"
	"github.com/nekomeowww/perobot/pkg/logger"
	twitter_public_types "github.com/nekomeowww/perobot/pkg/twitter/public/types"
	"go.uber.org/fx"
)

type NewModelParam struct {
	fx.In

	Logger        *logger.Logger
	TwitterPublic *thirdparty.TwitterPublic
}

type Model struct {
	Logger  *logger.Logger
	twitter *thirdparty.TwitterPublic
}

func NewModel() func(param NewModelParam) *Model {
	return func(param NewModelParam) *Model {
		return &Model{
			Logger:  param.Logger,
			twitter: param.TwitterPublic,
		}
	}
}

func (m *Model) GetOneTweet(tweetID string) (*twitter_public_types.TweetResultsResult, error) {
	tweetDetailResp, err := m.twitter.TweetDetail(tweetID)
	if err != nil {
		return nil, err
	}
	if tweetDetailResp.Data.ThreadedConversationWithInjectionsV2 == nil {
		m.Logger.WithField("tweet_id", tweetID).Warn("Tweet not found, threaded_conversation_with_injections_v2 is nil")
		return nil, nil
	}

	tweet := tweetDetailResp.Data.ThreadedConversationWithInjectionsV2.FindOneTweetEntry()
	if tweet == nil {
		m.Logger.WithField("tweet_id", tweetID).Warn("Tweet not found, tweet is nil")
		return nil, nil
	}

	tweetResult := tweet.TweetResults()
	if tweetResult == nil {
		m.Logger.WithField("tweet_id", tweetID).Warn("Tweet not found, tweet_result is nil")
		return nil, nil
	}

	return tweetResult, nil
}

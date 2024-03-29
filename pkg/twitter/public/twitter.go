package twitter_public

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/nekomeowww/perobot/pkg/options"
	twitter_public_types "github.com/nekomeowww/perobot/pkg/twitter/public/types"
)

type ClientOptions struct {
	Logger *logrus.Entry
}

func WithLogger(logger *logrus.Entry) options.CallOptions[ClientOptions] {
	return options.NewCallOptions(func(o *ClientOptions) {
		o.Logger = logger
	})
}

type Client struct {
	reqClient *req.Client

	logger               *logrus.Entry
	guestToken           string
	guestTokenObtainedAt time.Time
}

func NewClient(callOpts ...options.CallOptions[ClientOptions]) (*Client, error) {
	opts := options.ApplyCallOptions(callOpts, ClientOptions{
		Logger: logrus.NewEntry(logrus.New()),
	})

	c := req.
		C().
		SetBaseURL("https://api.twitter.com").
		SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.52").
		SetCommonHeader("Referer", "https://twitter.com/").
		SetCommonHeader("Origin", "https://twitter.com").
		SetCommonBearerAuthToken(twitter_public_types.GuestBearerToken)
	client := &Client{
		reqClient: c,
		logger:    opts.Logger,
	}

	return client, nil
}

func (c *Client) ActivateGuest() error {
	c.reqClient.ClearCookies()

	var guestActivateResp twitter_public_types.GuestActivateResp

	_, _, err := lo.AttemptWithDelay(100, time.Second, func(index int, duration time.Duration) error {
		resp, err := c.reqClient.R().
			SetResult(&guestActivateResp).
			Post("/1.1/guest/activate.json")
		if err != nil {
			c.logger.WithError(err).Error("failed to activate Twitter guest token, retrying...")
			return err
		}
		if !resp.IsSuccess() {
			c.logger.Error("failed to activate Twitter guest token, retrying...")
			return fmt.Errorf("request to %s failed: status code: %d", resp.Request.URL, resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		return err
	}

	c.guestToken = guestActivateResp.GuestToken
	c.guestTokenObtainedAt = time.Now()
	return nil
}

// TweetDetail 返回推文详情
//
// https://github.com/fa0311/TwitterInternalAPIDocument/blob/master/docs/markdown/GraphQL.md#tweetdetail
func (c *Client) TweetDetail(tweetID string) (*twitter_public_types.TweetDetailResp, error) {
	if c.guestTokenObtainedAt.IsZero() || time.Since(c.guestTokenObtainedAt) > 15*time.Minute {
		err := c.ActivateGuest()
		if err != nil {
			return nil, err
		}
	}

	newParamVariables := twitter_public_types.DefaultGetTweetDetailParamVariables
	newParamVariables.FocalTweetID = tweetID
	newParamVariablesJSON := string(lo.Must(json.Marshal(newParamVariables)))

	newParamFeatures := twitter_public_types.DefaultGetTweetDetailParamFeatures
	newParamFeaturesJSON := string(lo.Must(json.Marshal(newParamFeatures)))

	var tweetDetailResp twitter_public_types.TweetDetailResp
	resp, err := c.reqClient.R().
		SetQueryParam("variables", newParamVariablesJSON).
		SetQueryParam("features", newParamFeaturesJSON).
		SetHeader("X-Guest-Token", c.guestToken).
		SetResult(&tweetDetailResp).
		Get("/graphql/HQ_gjq7zDNvSiJOCSkwUEw/TweetDetail")
	if err != nil {
		c.logger.WithError(err).Error("failed to get tweet detail")
		return nil, err
	}
	if !resp.IsSuccess() {
		c.logger.Errorf("failed to get tweet detail, status code: %d, full request: %s", resp.StatusCode, resp.Dump())
		return nil, fmt.Errorf("request to %s failed: status code: %d", resp.Request.URL, resp.StatusCode)
	}

	return &tweetDetailResp, nil
}

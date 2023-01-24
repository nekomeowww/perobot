package tweet2images

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"github.com/nekomeowww/perobot/internal/models/twitter"
	"github.com/nekomeowww/perobot/pkg/handler"
	"github.com/nekomeowww/perobot/pkg/logger"
	twitter_public_types "github.com/nekomeowww/perobot/pkg/twitter/public/types"
)

type NewHandlerParam struct {
	fx.In

	Logger       *logger.Logger
	TwitterModel *twitter.Model
}

type Handler struct {
	Exchange sync.Map

	Logger  *logger.Logger
	Twitter *twitter.Model

	ReqClient *req.Client
}

func NewHandler() func(param NewHandlerParam) *Handler {
	return func(param NewHandlerParam) *Handler {
		handler := &Handler{
			Logger:    param.Logger,
			Twitter:   param.TwitterModel,
			ReqClient: req.C(),
		}
		return handler
	}
}

func (h *Handler) HandleChannelPostTweetToImages(c *handler.Context) {
	// 转发的消息不处理
	if c.Update.ChannelPost.ForwardFrom != nil {
		return
	}
	// 转发的消息不处理
	if c.Update.ChannelPost.ForwardFromChat != nil {
		return
	}

	tweetURL, err := url.Parse(c.Update.ChannelPost.Text)
	if err != nil {
		return
	}

	tweetRawURL := fmt.Sprintf("%s://%s%s", tweetURL.Scheme, tweetURL.Host, tweetURL.Path)
	tweetID := TweetIDFromText(tweetRawURL)
	if tweetID == "" {
		return
	}

	loggerFields := logrus.Fields{
		"tweet_id":   tweetID,
		"tweet_url":  tweetRawURL,
		"chat_id":    c.Update.ChannelPost.Chat.ID,
		"chat_title": c.Update.ChannelPost.Chat.Title,
	}

	var tweet *twitter_public_types.TweetResultsResult
	_, _, err = lo.AttemptWithDelay(10, time.Second, func(index int, duration time.Duration) error {
		tweet, err = h.Twitter.GetOneTweet(tweetID)
		if err != nil {
			return err
		}
		if tweet == nil {
			return nil
		}

		return nil
	})
	if err != nil {
		h.Logger.WithFields(loggerFields).Errorf("failed to get tweet, err: %v", err)
		return
	}
	if tweet == nil {
		h.Logger.WithFields(loggerFields).Warn("tweet not found")
		return
	}

	imageLinks := tweet.ExtendedPhotoURLs()
	if len(imageLinks) == 0 {
		h.Logger.WithField("tweet_id", tweetID).Warn("no images found in tweet, if tweet does contain images, then it is probably because the image contains adult content")
		return
	}

	h.Logger.WithFields(loggerFields).Info("tweet found, fetching images...")
	originalImagesLinks := lo.Map(imageLinks, func(link string, _ int) string { return tweetImageTo4kImage(link) })
	images := make([]*bytes.Buffer, 0, len(originalImagesLinks))
	for _, link := range originalImagesLinks {
		imageBuffer, err := h.fetchTweetImage(link, loggerFields)
		if err != nil {
			continue
		}

		images = append(images, imageBuffer)
	}
	if len(images) == 0 {
		h.Logger.WithFields(loggerFields).Warn("no images fetched, probably because of rate limit")
		return
	}

	h.Logger.WithFields(loggerFields).Infof("%d images fetched, sending to telegram...", len(images))

	tweetAuthor := tweet.User()
	var tweetAuthorInfo string
	if tweetAuthor == nil {
		tweetAuthorInfo = "未知"
	} else {
		tweetAuthorInfo = fmt.Sprintf(`<a href="https://twitter.com/%s">%s (@%s)</a>`, tweetAuthor.ScreenName, tweetAuthor.Name, tweetAuthor.ScreenName)
	}

	tweetContentInMarkdown := tweet.DisplayTextWithURLsMappedEmbeddedInHTML()
	if tweetContentInMarkdown != "" {
		tweetContentInMarkdown = "：\n\n" + tweetContentInMarkdown
	}

	mediaGroupConfig := tgbotapi.MediaGroupConfig{
		ChatID: c.Update.ChannelPost.Chat.ID,
		Media:  make([]interface{}, 0, len(images)),
	}
	for i, image := range images {
		file := tgbotapi.FileBytes{
			Name:  fmt.Sprintf("%s-%s", tweetID, filepath.Base(imageLinks[i])),
			Bytes: image.Bytes(),
		}

		inputMediaPhoto := tgbotapi.NewInputMediaPhoto(file)
		if i == 0 {
			inputMediaPhoto.ParseMode = "HTML"
			inputMediaPhoto.Caption = fmt.Sprintf(`%s%s`+"\n\n"+`来自 <a href="%s">Twitter</a>`,
				tweetAuthorInfo,
				tweetContentInMarkdown,
				tweetRawURL,
			)
			if inputMediaPhoto.Caption == "" {
				inputMediaPhoto.Caption = c.Update.ChannelPost.Text
			}

			h.Logger.Debugf("created a new input media photo with name: %s, size: %d, and caption: %s", file.Name, len(file.Bytes), inputMediaPhoto.Caption)
		} else {
			h.Logger.Debugf("created a new input media photo with name: %s, and size: %d", file.Name, len(file.Bytes))
		}

		mediaGroupConfig.Media = append(mediaGroupConfig.Media, inputMediaPhoto)
	}

	messages, err := c.Bot.SendMediaGroup(mediaGroupConfig)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	h.assignExchanges(messages[0].Chat.ID, messages[0].MessageID, tweetID, tweetAuthor.ScreenName, images, imageLinks)
	h.Logger.WithFields(loggerFields).Infof("%d images sent to channel", len(images))

	// 删除原始推文
	_, err = c.Bot.Request(tgbotapi.NewDeleteMessage(c.Update.ChannelPost.Chat.ID, c.Update.ChannelPost.MessageID))
	if err != nil {
		h.Logger.Error(err)
		return
	}
}

func (h *Handler) assignExchanges(chatID int64, messageID int, tweetID string, author string, images []*bytes.Buffer, imageLinks []string) {
	baseKey := fmt.Sprintf("key/tweet/%d/%d", chatID, messageID)
	h.Exchange.Store(baseKey, tweetID)
	h.Exchange.Store(baseKey+"/author", author)
	h.Exchange.Store(baseKey+"/images", images)
	h.Exchange.Store(baseKey+"/images/links", imageLinks)
}

func (h *Handler) cleanupExchanges(chatID int64, messageID int) {
	baseKey := fmt.Sprintf("key/tweet/%d/%d", chatID, messageID)
	h.Exchange.Delete(baseKey)
	h.Exchange.Delete(baseKey + "/author")
	h.Exchange.Delete(baseKey + "/images")
	h.Exchange.Delete(baseKey + "/images/links")
	h.Exchange.Delete(baseKey + "/processing")
}

var (
	TweetLinkIDRegexp = regexp.MustCompile(`https://twitter.com/([^/]+)/status/(\d+)`)
)

func TweetIDFromText(text string) string {
	matches := TweetLinkIDRegexp.FindStringSubmatch(text)
	if len(matches) != 3 {
		return ""
	}

	return matches[2]
}

// tweetImageTo4kImage 将推文中的图片链接转换为 4096x4096 的图片链接
func tweetImageTo4kImage(imageLink string) string {
	ext := filepath.Ext(imageLink)
	linkWithoutExt := strings.TrimSuffix(imageLink, ext)
	return fmt.Sprintf("%s?format=%s&name=4096x4096", linkWithoutExt, strings.TrimPrefix(ext, "."))
}

func (h *Handler) fetchTweetImage(link string, loggerFields logrus.Fields) (*bytes.Buffer, error) {
	loggerFields["image_url"] = link
	defer delete(loggerFields, "image_url")

	h.Logger.WithFields(loggerFields).Debugf("fetching image from tweet")

	buffer := new(bytes.Buffer)
	resp, err := h.ReqClient.R().SetOutput(buffer).Get(link)
	if err != nil {
		h.Logger.WithFields(loggerFields).Errorf("failed to fetch image from tweet, err: %v", err)
		return nil, err
	}
	if !resp.IsSuccess() {
		loggerFields["status_code"] = resp.StatusCode
		h.Logger.WithFields(loggerFields).Error("failed to fetch image from tweet")
		delete(loggerFields, "status_code")
		return nil, errors.New("failed to fetch image from tweet")
	}

	h.Logger.WithFields(loggerFields).Debugf("fetched image from tweet")
	return buffer, nil
}

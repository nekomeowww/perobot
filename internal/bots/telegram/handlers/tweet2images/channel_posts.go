package tweet2images

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"

	"github.com/nekomeowww/elapsing"
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

type FetchedTweetMedia struct {
	Type         twitter_public_types.EntityMediaType
	URL          string
	Body         *bytes.Buffer
	OriginalBody *bytes.Buffer
	Height       int
	Width        int
}

func (h *Handler) fetchImageMediaAsFetchedTweetMedia(
	media *twitter_public_types.ExtendedEntityMedia,
	logEntry *logrus.Entry,
	fc *elapsing.FuncCall,
) *FetchedTweetMedia {
	defer fc.Return()

	if media.MediaURLHTTPS == "" {
		return nil
	}

	regularURL := media.MediaURLHTTPS
	originalURL := tweetImageTo4kImage(regularURL)

	var regularImageBuffer *bytes.Buffer
	var originalImageBuffer *bytes.Buffer

	fc.StepEnds(elapsing.WithName("Generate regular and original image URLs"))

	wg := conc.NewWaitGroup()
	wg.Go(func() {
		var err error
		regularImageBuffer, err = h.fetchTweetMedia(regularURL, logEntry)
		if err != nil {
			logEntry.Errorf("failed to fetch regular images, err: %v", err)
		}
	})
	wg.Go(func() {
		var err error
		originalImageBuffer, err = h.fetchTweetMedia(originalURL, logEntry)
		if err != nil {
			logEntry.Errorf("failed to fetch original images, err: %v", err)
		}
	})

	wg.Wait()
	if regularImageBuffer == nil || originalImageBuffer == nil {
		return nil
	}

	fc.StepEnds(elapsing.WithName("Fetch regular and original images"))

	return &FetchedTweetMedia{
		Type:         twitter_public_types.TweetLegacyExtendedEntityMediaTypePhoto,
		URL:          regularURL,
		Body:         regularImageBuffer,
		OriginalBody: originalImageBuffer,
	}
}

func (h *Handler) fetchVideoMediaAsFetchedTweetMedia(
	media *twitter_public_types.ExtendedEntityMedia,
	logEntry *logrus.Entry,
	fc *elapsing.FuncCall,
) *FetchedTweetMedia {
	defer fc.Return()

	if media.VideoInfo == nil {
		return nil
	}
	if len(media.VideoInfo.Variants) == 0 {
		return nil
	}

	sort.SliceStable(media.VideoInfo.Variants, func(i, j int) bool {
		return media.VideoInfo.Variants[i].Bitrate > media.VideoInfo.Variants[j].Bitrate
	})

	regularURL := media.VideoInfo.Variants[0].URL
	originalURL := media.VideoInfo.Variants[len(media.VideoInfo.Variants)-1].URL
	fc.StepEnds(elapsing.WithName("Generate regular and original video URLs"))

	var regularVideoBuffer *bytes.Buffer
	var originalVideoBuffer *bytes.Buffer

	wg := conc.NewWaitGroup()
	wg.Go(func() {
		var err error
		regularVideoBuffer, err = h.fetchTweetMedia(regularURL, logEntry)
		if err != nil {
			logEntry.Errorf("failed to fetch regular videos, err: %v", err)
		}
	})
	wg.Go(func() {
		var err error
		originalVideoBuffer, err = h.fetchTweetMedia(originalURL, logEntry)
		if err != nil {
			logEntry.Errorf("failed to fetch original videos, err: %v", err)
		}
	})

	wg.Wait()
	if regularVideoBuffer == nil || originalVideoBuffer == nil {
		return nil
	}

	fc.StepEnds(elapsing.WithName("Fetch regular and original videos"))

	return &FetchedTweetMedia{
		Type:         twitter_public_types.TweetLegacyExtendedEntityMediaTypeVideo,
		URL:          regularURL,
		Body:         regularVideoBuffer,
		OriginalBody: originalVideoBuffer,
		Height:       media.Sizes.Large.H,
		Width:        media.Sizes.Large.W,
	}
}

func (h *Handler) newFetchingImageWorkerFunction(
	mediasSlice []*FetchedTweetMedia,
	mediaSliceIndex int,
	media *twitter_public_types.ExtendedEntityMedia,
	logEntry *logrus.Entry,
	fc *elapsing.FuncCall,
) func() {
	return func() {
		defer fc.Return()

		fetchedTweetMedia := h.fetchImageMediaAsFetchedTweetMedia(media, logEntry, fc.ForFunc())
		if fetchedTweetMedia != nil {
			mediasSlice[mediaSliceIndex] = fetchedTweetMedia
		} else {
			logEntry.Warnf("failed to fetch image: %s", media.URL)
		}
	}
}

func (h *Handler) newFetchingVideoWorkerFunction(
	mediasSlice []*FetchedTweetMedia,
	mediaSliceIndex int,
	media *twitter_public_types.ExtendedEntityMedia,
	logEntry *logrus.Entry,
	fc *elapsing.FuncCall,
) func() {
	return func() {
		defer fc.Return()

		fetchedTweetMedia := h.fetchVideoMediaAsFetchedTweetMedia(media, logEntry, fc.ForFunc())
		if fetchedTweetMedia != nil {
			mediasSlice[mediaSliceIndex] = fetchedTweetMedia
		} else {
			logEntry.Warnf("failed to fetch video: %s", media.URL)
		}
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

	e := elapsing.New()
	tweetURL, err := url.Parse(c.Update.ChannelPost.Text)
	if err != nil {
		return
	}
	e.StepEnds(elapsing.WithName("Parse URL"))

	tweetRawURL := fmt.Sprintf("%s://%s%s", tweetURL.Scheme, tweetURL.Host, tweetURL.Path)
	tweetID := TweetIDFromText(tweetRawURL)
	if tweetID == "" {
		return
	}
	e.StepEnds(elapsing.WithName("Extract Tweet ID"))

	logEntry := h.Logger.WithFields(logrus.Fields{
		"tweet_id":   tweetID,
		"tweet_url":  tweetRawURL,
		"chat_id":    c.Update.ChannelPost.Chat.ID,
		"chat_title": c.Update.ChannelPost.Chat.Title,
	})

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
		logEntry.Errorf("failed to get tweet, err: %v", err)
		return
	}
	if tweet == nil {
		logEntry.Warn("tweet not found")
		return
	}
	e.StepEnds(elapsing.WithName("Fetch TweetDetail"))

	medias := tweet.ExtendedMedias()
	if len(medias) == 0 {
		h.Logger.WithField("tweet_id", tweetID).Warn("no images/videos found in tweet, if tweet does contain images, then it is probably because the image contains adult content")
		return
	}

	e.StepEnds(elapsing.WithName("Extract Tweet Medias"))
	medias = lo.Filter(medias, func(item *twitter_public_types.ExtendedEntityMedia, _ int) bool {
		return lo.Contains([]twitter_public_types.EntityMediaType{
			twitter_public_types.TweetLegacyExtendedEntityMediaTypePhoto,
			twitter_public_types.TweetLegacyExtendedEntityMediaTypeVideo,
			twitter_public_types.TweetLegacyExtendedEntityMediaTypeAnimatedGIF,
		}, item.Type)
	})

	logEntry.Infof("tweet found, fetching %d images/videos...", len(medias))

	wg := conc.NewWaitGroup()
	fetchedMedias := make([]*FetchedTweetMedia, len(medias))
	for i, media := range medias {
		switch media.Type {
		case twitter_public_types.TweetLegacyExtendedEntityMediaTypePhoto:
			wg.Go(h.newFetchingImageWorkerFunction(fetchedMedias, i, media, logEntry, e.ForFunc()))
		case twitter_public_types.TweetLegacyExtendedEntityMediaTypeVideo:
			wg.Go(h.newFetchingVideoWorkerFunction(fetchedMedias, i, media, logEntry, e.ForFunc()))
		case twitter_public_types.TweetLegacyExtendedEntityMediaTypeAnimatedGIF:
			wg.Go(h.newFetchingVideoWorkerFunction(fetchedMedias, i, media, logEntry, e.ForFunc()))
		default:
		}
	}

	wg.Wait()
	fetchedMedias = lo.Filter(fetchedMedias, func(item *FetchedTweetMedia, _ int) bool { return item != nil })
	if len(fetchedMedias) == 0 {
		logEntry.Warn("no images/videos fetched, probably because of rate limit")
		return
	}

	logEntry.Infof("%d images/videos fetched, sending to telegram...", len(fetchedMedias))
	e.StepEnds(elapsing.WithName("Fetch Medias"))

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
	e.StepEnds(elapsing.WithName("Construct Message Content"))

	mediaGroupConfig := tgbotapi.MediaGroupConfig{
		ChatID: c.Update.ChannelPost.Chat.ID,
		Media:  make([]interface{}, 0, len(fetchedMedias)),
	}
	for i, media := range fetchedMedias {
		file := tgbotapi.FileBytes{
			Name:  fmt.Sprintf("%s-%s", tweetID, filepath.Base(media.URL)),
			Bytes: media.Body.Bytes(),
		}

		caption := fmt.Sprintf(`%s%s`+"\n\n"+`来自 <a href="%s">Twitter</a>`,
			tweetAuthorInfo,
			tweetContentInMarkdown,
			tweetRawURL,
		)

		switch media.Type {
		case twitter_public_types.TweetLegacyExtendedEntityMediaTypePhoto:
			inputMediaPhoto := tgbotapi.NewInputMediaPhoto(file)
			if i == 0 {
				inputMediaPhoto.ParseMode = "HTML"
				inputMediaPhoto.Caption = caption
				if inputMediaPhoto.Caption == "" {
					inputMediaPhoto.Caption = c.Update.ChannelPost.Text
				}

				h.Logger.Debugf("created a new input media photo with name: %s, size: %d, and caption: %s", file.Name, len(file.Bytes), inputMediaPhoto.Caption)
			} else {
				h.Logger.Debugf("created a new input media photo with name: %s, and size: %d", file.Name, len(file.Bytes))
			}

			mediaGroupConfig.Media = append(mediaGroupConfig.Media, inputMediaPhoto)
		case twitter_public_types.TweetLegacyExtendedEntityMediaTypeVideo:
			inputMediaVideo := tgbotapi.NewInputMediaVideo(file)
			if i == 0 {
				inputMediaVideo.ParseMode = "HTML"
				inputMediaVideo.Caption = caption
				inputMediaVideo.Height = media.Height
				inputMediaVideo.Width = media.Width
				if inputMediaVideo.Caption == "" {
					inputMediaVideo.Caption = c.Update.ChannelPost.Text
				}

				h.Logger.Debugf("created a new input media video with name: %s, size: %d, and caption: %s", file.Name, len(file.Bytes), inputMediaVideo.Caption)
			} else {
				h.Logger.Debugf("created a new input media video with name: %s, and size: %d", file.Name, len(file.Bytes))
			}

			mediaGroupConfig.Media = append(mediaGroupConfig.Media, inputMediaVideo)
		}
	}

	e.StepEnds(elapsing.WithName("Construct MediaGroupConfig"))

	messages, err := c.Bot.SendMediaGroup(mediaGroupConfig)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	e.StepEnds(elapsing.WithName("Send MediaGroup"))

	h.assignExchanges(messages[0].Chat.ID, messages[0].MessageID, tweetID, tweetAuthor.ScreenName, fetchedMedias)
	logEntry.Infof("%d images/videos sent to channel", len(fetchedMedias))

	e.StepEnds(elapsing.WithName("Assign Exchanges"))

	// 删除原始推文
	_, err = c.Bot.Request(tgbotapi.NewDeleteMessage(c.Update.ChannelPost.Chat.ID, c.Update.ChannelPost.MessageID))
	if err != nil {
		h.Logger.Error(err)
		return
	}

	e.StepEnds(elapsing.WithName("Delete Original Message"))
	go h.Logger.Debugf("Tweet to media done, time cost:\n%s", e.Stats())
}

func (h *Handler) assignExchanges(chatID int64, messageID int, tweetID string, author string, medias []*FetchedTweetMedia) {
	baseKey := fmt.Sprintf("key/tweet/%d/%d", chatID, messageID)
	h.Exchange.Store(baseKey, tweetID)
	h.Exchange.Store(baseKey+"/author", author)
	h.Exchange.Store(baseKey+"/medias", medias)
}

func (h *Handler) cleanupExchanges(chatID int64, messageID int) {
	baseKey := fmt.Sprintf("key/tweet/%d/%d", chatID, messageID)
	h.Exchange.Delete(baseKey)
	h.Exchange.Delete(baseKey + "/author")
	h.Exchange.Delete(baseKey + "/medias")
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

func (h *Handler) fetchTweetMedia(link string, logEntry *logrus.Entry) (*bytes.Buffer, error) {
	logEntry.WithField("image_url", link).Debugf("fetching image from tweet")

	buffer := new(bytes.Buffer)
	resp, err := h.ReqClient.R().SetOutput(buffer).Get(link)
	if err != nil {
		logEntry.WithField("image_url", link).Errorf("failed to fetch image from tweet, err: %v", err)
		return nil, err
	}
	if !resp.IsSuccess() {
		logEntry.WithFields(logrus.Fields{
			"image_url":   link,
			"status_code": resp.StatusCode,
		}).Error("failed to fetch image from tweet")
		return nil, errors.New("failed to fetch image from tweet")
	}

	logEntry.WithField("image_url", link).Debugf("fetched image from tweet")
	return buffer, nil
}

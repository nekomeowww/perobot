package pixiv2images

import (
	"bytes"
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

	"github.com/nekomeowww/perobot/internal/thirdparty"
	"github.com/nekomeowww/perobot/pkg/handler"
	"github.com/nekomeowww/perobot/pkg/logger"
	pixiv_public_types "github.com/nekomeowww/perobot/pkg/pixiv/public/types"
)

type NewHandlerParam struct {
	fx.In

	Logger *logger.Logger
	Pixiv  *thirdparty.PixivPublic
}

type Handler struct {
	Exchange sync.Map
	Logger   *logger.Logger
	Pixiv    *thirdparty.PixivPublic

	ReqClient *req.Client
}

func NewHandler() func(param NewHandlerParam) *Handler {
	return func(param NewHandlerParam) *Handler {
		handler := &Handler{
			Logger:    param.Logger,
			Pixiv:     param.Pixiv,
			ReqClient: req.C(),
		}
		return handler
	}
}

func (h *Handler) HandleChannelPostPixivToImages(c *handler.Context) {
	// 转发的消息不处理
	if c.Update.ChannelPost.ForwardFrom != nil {
		return
	}
	// 转发的消息不处理
	if c.Update.ChannelPost.ForwardFromChat != nil {
		return
	}

	_, err := c.Bot.Request(tgbotapi.ChatActionConfig{
		BaseChat: tgbotapi.BaseChat{ChatID: c.Update.ChannelPost.Chat.ID},
		Action:   "upload_photo",
	})
	if err != nil {
		h.Logger.Errorf("failed to send chat action, err: %v", err)
		// PASS
	}

	pixivIllustURL, err := url.Parse(c.Update.ChannelPost.Text)
	if err != nil {
		return
	}

	pixivIllustRawURL := fmt.Sprintf("%s://%s%s", pixivIllustURL.Scheme, pixivIllustURL.Host, pixivIllustURL.Path)
	illustID := IllustIDFromText(pixivIllustRawURL)
	if illustID == "" {
		return
	}

	loggerFields := logrus.Fields{
		"pixiv_illust_id":  illustID,
		"pixiv_illust_url": pixivIllustRawURL,
		"chat_id":          c.Update.ChannelPost.Chat.ID,
		"chat_title":       c.Update.ChannelPost.Chat.Title,
	}

	var illustDetailResp *pixiv_public_types.IllustDetailResp
	_, _, err = lo.AttemptWithDelay(1, time.Second, func(index int, duration time.Duration) error {
		illustDetailResp, err = h.Pixiv.IllustDetail(illustID)
		if err != nil {
			return err
		}
		if illustDetailResp == nil {
			return nil
		}

		return nil
	})
	if err != nil {
		h.Logger.WithFields(loggerFields).Errorf("failed to get pixiv illust detail, err: %v", err)
		return
	}
	if illustDetailResp == nil {
		h.Logger.WithFields(loggerFields).Warn("pixiv illust detail not found")
		return
	}
	if illustDetailResp.Body == nil {
		h.Logger.WithFields(loggerFields).Warn("pixiv illust detail body is nil")
		return
	}

	var illustDetailPagesResp *pixiv_public_types.IllustDetailPagesResp
	_, _, err = lo.AttemptWithDelay(1, time.Second, func(index int, duration time.Duration) error {
		illustDetailPagesResp, err = h.Pixiv.IllustDetailPages(illustID)
		if err != nil {
			return err
		}
		if illustDetailPagesResp == nil {
			return nil
		}

		return nil
	})
	if err != nil {
		h.Logger.WithFields(loggerFields).Errorf("failed to get pixiv illust detail pages, err: %v", err)
		return
	}
	if illustDetailPagesResp == nil {
		h.Logger.WithFields(loggerFields).Warn("pixiv illust detail pages not found")
		return
	}

	urlItems := lo.Filter(illustDetailPagesResp.Body, func(item *pixiv_public_types.IllustDetailPagesRespItem, _ int) bool {
		return item.Urls.Regular != "" && item.Urls.Original != ""
	})

	regularURLs := lo.Map(urlItems, func(item *pixiv_public_types.IllustDetailPagesRespItem, _ int) string { return item.Urls.Regular })
	originalURLs := lo.Map(urlItems, func(item *pixiv_public_types.IllustDetailPagesRespItem, _ int) string { return item.Urls.Original })
	if len(regularURLs) == 0 || len(originalURLs) == 0 {
		h.Logger.WithFields(loggerFields).Warn("no image found")
		return
	}

	regularURLs = lo.Slice(regularURLs, 0, 4)
	originalURLs = lo.Slice(originalURLs, 0, 4)

	regularImages := make([]*bytes.Buffer, 0, len(regularURLs))
	originalImages := make([]*bytes.Buffer, 0, len(originalURLs))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		h.Logger.WithFields(loggerFields).Info("fetching regular images...")

		for _, url := range regularURLs {
			imageBuffer, err := h.fetchPixivIllustImage(url, loggerFields)
			if err != nil {
				continue
			}

			regularImages = append(regularImages, imageBuffer)
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		h.Logger.WithFields(loggerFields).Info("fetching original images...")

		for _, url := range originalURLs {
			imageBuffer, err := h.fetchPixivIllustImage(url, loggerFields)
			if err != nil {
				continue
			}

			originalImages = append(originalImages, imageBuffer)
		}

		wg.Done()
	}()

	wg.Wait()
	h.Logger.WithFields(loggerFields).Infof("%d regular images, %d original images fetched, sending to telegram...", len(regularImages), len(originalImages))

	var illustAuthorInfo string
	if illustDetailResp.Body.UserName == "" {
		illustAuthorInfo = "未知"
	} else {
		illustAuthorInfo = fmt.Sprintf(`<a href="https://www.pixiv.net/users/%s">%s</a>`, illustDetailResp.Body.UserID, illustDetailResp.Body.UserName)
	}

	illustContentInMarkdown := illustDetailResp.Body.Title
	if illustContentInMarkdown != "" {
		illustContentInMarkdown = "：\n\n" + illustContentInMarkdown
	}

	// 写入标签
	tags := make([]string, 0, len(illustDetailResp.Body.Tags.Tags))
	for _, tag := range illustDetailResp.Body.Tags.Tags {
		tags = append(tags, fmt.Sprintf("#%s", tag.Tag))
	}
	illustContentInMarkdown += strings.Join(tags, " ")

	mediaGroupConfig := tgbotapi.MediaGroupConfig{
		ChatID: c.Update.ChannelPost.Chat.ID,
		Media:  make([]interface{}, 0, len(regularImages)),
	}
	for i, image := range regularImages {
		file := tgbotapi.FileBytes{
			Name:  fmt.Sprintf("%s-%s", illustID, filepath.Base(regularURLs[i])),
			Bytes: image.Bytes(),
		}

		inputMediaPhoto := tgbotapi.NewInputMediaPhoto(file)
		if i == 0 {
			inputMediaPhoto.ParseMode = "HTML"
			inputMediaPhoto.Caption = fmt.Sprintf(`%s%s`+"\n\n"+`来自 <a href="%s">Pixiv</a>`,
				illustAuthorInfo,
				illustContentInMarkdown,
				pixivIllustRawURL,
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

	h.assignExchanges(messages[0].Chat.ID, messages[0].MessageID, illustID, illustDetailResp.Body.UserName, regularImages, originalImages, regularURLs)
	h.Logger.WithFields(loggerFields).Infof("%d images sent to channel", len(regularImages))

	// 删除原始 Pixiv 消息
	_, err = c.Bot.Request(tgbotapi.NewDeleteMessage(c.Update.ChannelPost.Chat.ID, c.Update.ChannelPost.MessageID))
	if err != nil {
		h.Logger.Error(err)
		return
	}
}

func (h *Handler) assignExchanges(
	chatID int64,
	messageID int,
	illustID string,
	author string,
	regularImages []*bytes.Buffer,
	originalImages []*bytes.Buffer,
	urls []string,
) {
	baseKey := fmt.Sprintf("key/pixiv/%d/%d", chatID, messageID)
	h.Exchange.Store(baseKey, illustID)
	h.Exchange.Store(baseKey+"/author", author)
	h.Exchange.Store(baseKey+"/images/regular", regularImages)
	h.Exchange.Store(baseKey+"/images/original", originalImages)
	h.Exchange.Store(baseKey+"/images/urls", urls)
}

func (h *Handler) cleanupExchanges(chatID int64, messageID int) {
	baseKey := fmt.Sprintf("key/pixiv/%d/%d", chatID, messageID)
	h.Exchange.Delete(baseKey)
	h.Exchange.Delete(baseKey + "/author")
	h.Exchange.Delete(baseKey + "/images/regular")
	h.Exchange.Delete(baseKey + "/images/original")
	h.Exchange.Delete(baseKey + "/images/urls")
	h.Exchange.Delete(baseKey + "/processing")
}

func (h *Handler) fetchPixivIllustImage(link string, loggerFields logrus.Fields) (*bytes.Buffer, error) {
	loggerFields["image_url"] = link
	defer delete(loggerFields, "image_url")

	h.Logger.WithFields(loggerFields).Debugf("fetching pixiv image")

	buffer, err := h.Pixiv.GetImage(link)
	if err != nil {
		h.Logger.WithFields(loggerFields).Errorf("failed to fetch pixiv image, err: %v", err)
		return nil, err
	}

	h.Logger.WithFields(loggerFields).Debugf("fetched pixiv image")
	return buffer, nil
}

var (
	PixivIllustIDRegexp = regexp.MustCompile(`https://www.pixiv.net/(.*\/)?artworks/(\d+)`)
)

func IllustIDFromText(text string) string {
	matches := PixivIllustIDRegexp.FindStringSubmatch(text)
	if len(matches) != 3 {
		return ""
	}

	return matches[2]
}

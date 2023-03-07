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
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"

	"github.com/nekomeowww/elapsing"
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

func (h *Handler) fetchImage(
	imageSlice []*bytes.Buffer,
	imageSliceIndex int,
	url string,
	logEntry *logrus.Entry,
	fc *elapsing.FuncCall,
) func() {
	return func() {
		defer fc.Return()

		imageBuffer, err := h.fetchPixivIllustImage(url, logEntry)
		if err != nil {
			return
		}

		imageSlice[imageSliceIndex] = imageBuffer
		fc.StepEnds(elapsing.WithName("Fetch Image"))
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

	e := elapsing.New()
	pixivIllustURL, err := url.Parse(c.Update.ChannelPost.Text)
	if err != nil {
		return
	}

	e.StepEnds(elapsing.WithName("Parse URL"))

	pixivIllustRawURL := fmt.Sprintf("%s://%s%s", pixivIllustURL.Scheme, pixivIllustURL.Host, pixivIllustURL.Path)
	illustID := IllustIDFromText(pixivIllustRawURL)
	if illustID == "" {
		return
	}

	e.StepEnds(elapsing.WithName("Extract Pixiv Illust ID"))

	loggerEntry := h.Logger.WithFields(logrus.Fields{
		"pixiv_illust_id":  illustID,
		"pixiv_illust_url": pixivIllustRawURL,
		"chat_id":          c.Update.ChannelPost.Chat.ID,
		"chat_title":       c.Update.ChannelPost.Chat.Title,
	})

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
		loggerEntry.Errorf("failed to get pixiv illust detail, err: %v", err)
		return
	}
	if illustDetailResp == nil {
		loggerEntry.Warn("pixiv illust detail not found")
		return
	}
	if illustDetailResp.Body == nil {
		loggerEntry.Warn("pixiv illust detail body is nil")
		return
	}
	e.StepEnds(elapsing.WithName("Get Pixiv Illust Detail"))

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
		loggerEntry.Errorf("failed to get pixiv illust detail pages, err: %v", err)
		return
	}
	if illustDetailPagesResp == nil {
		loggerEntry.Warn("pixiv illust detail pages not found")
		return
	}
	e.StepEnds(elapsing.WithName("Get Pixiv Illust Detail Pages"))

	urlItems := lo.Filter(illustDetailPagesResp.Body, func(item *pixiv_public_types.IllustDetailPagesRespItem, _ int) bool {
		return item.Urls.Regular != "" && item.Urls.Original != ""
	})

	regularURLs := lo.Map(urlItems, func(item *pixiv_public_types.IllustDetailPagesRespItem, _ int) string { return item.Urls.Regular })
	originalURLs := lo.Map(urlItems, func(item *pixiv_public_types.IllustDetailPagesRespItem, _ int) string { return item.Urls.Original })
	if len(regularURLs) == 0 || len(originalURLs) == 0 {
		loggerEntry.Warn("no image found")
		return
	}

	regularURLs = lo.Slice(regularURLs, 0, 4)
	originalURLs = lo.Slice(originalURLs, 0, 4)
	e.StepEnds(elapsing.WithName("Extract and filter Pixiv Illust Detail Pages"))

	regularImages := make([]*bytes.Buffer, len(regularURLs))
	originalImages := make([]*bytes.Buffer, len(originalURLs))

	wg := conc.NewWaitGroup()
	for i, url := range regularURLs {
		wg.Go(h.fetchImage(regularImages, i, url, loggerEntry, e.ForFunc()))
	}
	for i, url := range originalURLs {
		wg.Go(h.fetchImage(originalImages, i, url, loggerEntry, e.ForFunc()))
	}

	wg.Wait()
	loggerEntry.Infof("%d regular images, %d original images fetched, sending to telegram...", len(regularImages), len(originalImages))
	e.StepEnds(elapsing.WithName("Fetch Pixiv Illust Images"))

	regularImages = lo.Filter(regularImages, func(item *bytes.Buffer, _ int) bool { return item != nil })
	originalImages = lo.Filter(originalImages, func(item *bytes.Buffer, _ int) bool { return item != nil })
	if len(regularImages) == 0 || len(originalImages) == 0 {
		loggerEntry.Warn("no image can be fetched")
		return
	}

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
		tagStr := strings.ReplaceAll(tag.Tag, "-", "")
		tags = append(tags, fmt.Sprintf("#%s", tagStr))
	}
	illustContentInMarkdown += fmt.Sprintf("\n\n%s", strings.Join(tags, " "))
	e.StepEnds(elapsing.WithName("Build Pixiv Illust Content"))

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
	e.StepEnds(elapsing.WithName("Construct MediaGroupConfig"))

	messages, err := c.Bot.SendMediaGroup(mediaGroupConfig)
	if err != nil {
		h.Logger.Error(err)
		return
	}
	e.StepEnds(elapsing.WithName("Send MediaGroup"))

	h.assignExchanges(messages[0].Chat.ID, messages[0].MessageID, illustID, illustDetailResp.Body.UserName, regularImages, originalImages, regularURLs)
	loggerEntry.Infof("%d images sent to channel", len(regularImages))
	e.StepEnds(elapsing.WithName("Assign Exchanges"))

	// 删除原始 Pixiv 消息
	_, err = c.Bot.Request(tgbotapi.NewDeleteMessage(c.Update.ChannelPost.Chat.ID, c.Update.ChannelPost.MessageID))
	if err != nil {
		h.Logger.Error(err)
		return
	}
	e.StepEnds(elapsing.WithName("Delete Original Pixiv Message"))
	go h.Logger.Debugf("Pixiv to image done, time cost:\n%s", e.Stats())
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

func (h *Handler) fetchPixivIllustImage(link string, logEntry *logrus.Entry) (*bytes.Buffer, error) {
	logEntry.WithField("image_url", link).Debugf("fetching pixiv image")

	buffer, err := h.Pixiv.GetImage(link)
	if err != nil {
		logEntry.WithField("image_url", link).Errorf("failed to fetch pixiv image, err: %v", err)
		return nil, err
	}

	logEntry.WithField("image_url", link).Debugf("fetched pixiv image")
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

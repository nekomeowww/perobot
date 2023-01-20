package pixiv2images

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
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
	Logger *logger.Logger
	Pixiv  *thirdparty.PixivPublic

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
		return item.Urls.Original != ""
	})

	urls := lo.Map(urlItems, func(item *pixiv_public_types.IllustDetailPagesRespItem, _ int) string {
		return item.Urls.Original
	})

	urls = lo.Slice(urls, 0, 4)
	h.Logger.WithFields(loggerFields).Info("illust found, fetching images...")
	images := make([]*bytes.Buffer, 0, len(urls))
	for _, link := range urls {
		imageBuffer, err := h.fetchPixivIllustImage(link, loggerFields)
		if err != nil {
			continue
		}

		images = append(images, imageBuffer)
	}

	var illustAuthorInfo string
	if illustDetailResp.Body.UserName == "" {
		illustAuthorInfo = "未知"
	} else {
		illustAuthorInfo = fmt.Sprintf(`<a href="https://www.pixiv.net/users/%s">%s</a>`, illustDetailResp.Body.UserID, illustDetailResp.Body.UserName)
	}

	illustContentInMarkdown := illustDetailResp.Body.Title
	if illustContentInMarkdown != "" {
		illustContentInMarkdown += "\n\n"
	}

	illustContentInMarkdown = strings.ReplaceAll(illustContentInMarkdown, "<br />", "\n")

	h.Logger.WithFields(loggerFields).Info("images fetched, sending to telegram...")
	mediaGroupConfig := tgbotapi.MediaGroupConfig{
		ChatID: c.Update.ChannelPost.Chat.ID,
		Media:  make([]interface{}, 0, len(images)),
	}
	for i, image := range images {
		inputMediaPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FileBytes{
			Name:  fmt.Sprintf("%s-%s", illustID, filepath.Base(urls[i])),
			Bytes: image.Bytes(),
		})
		if i == 0 {
			inputMediaPhoto.ParseMode = "HTML"
			inputMediaPhoto.Caption = fmt.Sprintf(`%sBy: %s`+"\n\n"+`<a href="%s">Source</a>`,
				illustContentInMarkdown,
				illustAuthorInfo,
				pixivIllustRawURL,
			)
			if inputMediaPhoto.Caption == "" {
				inputMediaPhoto.Caption = c.Update.ChannelPost.Text
			}

			h.Logger.Debugf("new images message with caption: %s", inputMediaPhoto.Caption)
		}

		mediaGroupConfig.Media = append(mediaGroupConfig.Media, inputMediaPhoto)
	}

	_, err = c.Bot.SendMediaGroup(mediaGroupConfig)
	if err != nil {
		h.Logger.Error(err)
		return
	}

	h.Logger.WithFields(loggerFields).Info("images sent to telegram")

	// 删除原始 Pixiv 消息
	_, err = c.Bot.Request(tgbotapi.NewDeleteMessage(c.Update.ChannelPost.Chat.ID, c.Update.ChannelPost.MessageID))
	if err != nil {
		h.Logger.Error(err)
		return
	}
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

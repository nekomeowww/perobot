package pixiv2images

import (
	"bytes"
	"fmt"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/nekomeowww/perobot/pkg/handler"
)

func (h *Handler) HandleMessageAutomaticForwardedFromLinkedChannel(c *handler.Context) {
	// 非自动转发的消息不处理
	if !c.Update.Message.IsAutomaticForward {
		return
	}
	// 不是转发的消息不处理
	if c.Update.Message.ForwardFromChat == nil {
		return
	}

	// 等待 ChannelPostPixivToImages 处理完毕并获取到 chat id 和 message id
	time.Sleep(time.Second)

	baseKey := fmt.Sprintf("key/pixiv/%d/%d", c.Update.Message.ForwardFromChat.ID, c.Update.Message.ForwardFromMessageID)
	illustID, ok := h.Exchange.Load(baseKey)
	if !ok {
		return
	}

	defer h.cleanupExchanges(c.Update.Message.ForwardFromChat.ID, c.Update.Message.ForwardFromMessageID)

	illustIDFilesPostingProcessing, ok := h.Exchange.Load(baseKey + "/processing")
	if ok && illustIDFilesPostingProcessing == true {
		// 有可能正在处理中，去重
		return
	}

	h.Exchange.Store(baseKey+"/processing", true)

	loggerFields := logrus.Fields{
		"chat_id":                 c.Update.Message.Chat.ID,
		"chat_title":              c.Update.Message.Chat.Title,
		"forward_from_chat_id":    c.Update.Message.ForwardFromChat.ID,
		"forward_from_chat_title": c.Update.Message.ForwardFromChat.Title,
		"forward_from_message_id": c.Update.Message.ForwardFromMessageID,
		"tweet_id":                illustID,
	}

	author, ok := h.Exchange.Load(baseKey + "/author")
	if !ok {
		h.Logger.WithFields(loggerFields).Error("author not found")
		return
	}

	authorName, ok := author.(string)
	if !ok {
		h.Logger.WithFields(loggerFields).Error("author not found, incorrect type, type is not string")
		return
	}

	imagesRaw, ok := h.Exchange.Load(baseKey + "/images/original")
	if !ok {
		h.Logger.WithFields(loggerFields).Error("images not found")
		return
	}

	images, ok := imagesRaw.([]*bytes.Buffer)
	if !ok {
		h.Logger.WithFields(loggerFields).Error("images not found, incorrect type, type is not []*bytes.Buffer")
		return
	}

	imageLinksRaw, ok := h.Exchange.Load(baseKey + "/images/urls")
	if !ok {
		h.Logger.WithFields(loggerFields).Error("image links not found")
		return
	}

	imageLinks, ok := imageLinksRaw.([]string)
	if !ok {
		h.Logger.WithFields(loggerFields).Error("image url not found, incorrect type, type is not []string")
		return
	}

	botChatMemberInOriginalChannel, err := c.Bot.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: tgbotapi.ChatConfigWithUser{ChatID: c.Update.Message.ForwardFromChat.ID, UserID: c.Bot.Self.ID}})
	if err != nil {
		h.Logger.WithFields(loggerFields).Error(err)
		return
	}

	if botChatMemberInOriginalChannel.Status != "administrator" {
		h.Logger.WithFields(loggerFields).Warn("received a message from a channel that the bot is not an administrator in, ignoring...")
		return
	}

	h.Logger.Info("linked channel message received, processing... prepare to send images to discussion group")
	botChatMemberInDiscussionGroup, err := c.Bot.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: tgbotapi.ChatConfigWithUser{ChatID: c.Update.Message.Chat.ID, UserID: c.Bot.Self.ID}})
	if err != nil {
		h.Logger.WithFields(loggerFields).Error(err)
		return
	}
	if botChatMemberInDiscussionGroup.Status != "administrator" && !botChatMemberInDiscussionGroup.CanSendMediaMessages {
		h.Logger.WithFields(loggerFields).Error("bot is not an administrator in the discussion group or does not have the permission to send messages or media messages")
		return
	}

	h.Logger.Info("generating thumbnails...")
	thumbnailImages := make([]*bytes.Buffer, len(images))
	for i, image := range images {
		img, err := imaging.Decode(bytes.NewReader(image.Bytes()))
		if err != nil {
			h.Logger.WithFields(loggerFields).Error(err)
			continue
		}

		newImg := imaging.Resize(img, 320, 0, imaging.Lanczos)
		if newImg.Rect.Dy() > 320 {
			newImg = imaging.CropCenter(newImg, 320, 320)
		}

		thumbnailImages[i] = new(bytes.Buffer)
		err = imaging.Encode(thumbnailImages[i], newImg, imaging.JPEG)
		if err != nil {
			h.Logger.WithFields(loggerFields).Error(err)
			continue
		}
	}

	h.Logger.Info("sending images to discussion group...")
	mediaGroupConfig := tgbotapi.MediaGroupConfig{
		ReplyToMessageID: c.Update.Message.MessageID,
		ChatID:           c.Update.Message.Chat.ID,
		Media:            make([]interface{}, 0, len(images)),
	}

	for i, image := range images {
		file := tgbotapi.FileBytes{
			Name: fmt.Sprintf("pixiv-by-%s-%s-%s",
				authorName,
				illustID,
				filepath.Base(imageLinks[i]),
			),
			Bytes: image.Bytes(),
		}

		inputMediaDocument := tgbotapi.NewInputMediaDocument(file)
		h.Logger.Debugf("created a new input media document with name: %s, and size: %d", file.Name, len(file.Bytes))

		if thumbnailImages[i] != nil {
			thumbFile := tgbotapi.FileBytes{
				Name:  "thumbnail-" + file.Name,
				Bytes: thumbnailImages[i].Bytes(),
			}

			inputMediaDocument.Thumb = thumbFile
			h.Logger.Debugf("created a new input media document thumbnail with name: %s, and size: %d", thumbFile.Name, len(thumbFile.Bytes))
		}

		mediaGroupConfig.Media = append(mediaGroupConfig.Media, inputMediaDocument)
	}

	_, err = c.Bot.SendMediaGroup(mediaGroupConfig)
	if err != nil {
		h.Logger.WithFields(loggerFields).Error(err)
		return
	}

	h.Logger.WithFields(loggerFields).Infof("%d images sent as comment of channel post in discussion group", len(images))

	for _, i := range images {
		i.Reset()
	}
}

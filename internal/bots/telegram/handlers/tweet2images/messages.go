package tweet2images

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/disintegration/imaging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/nekomeowww/perobot/pkg/handler"
	twitter_public_types "github.com/nekomeowww/perobot/pkg/twitter/public/types"
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

	baseKey := fmt.Sprintf("key/tweet/%d/%d", c.Update.Message.ForwardFromChat.ID, c.Update.Message.ForwardFromMessageID)

	tweetID, ok := h.Exchange.Load(baseKey)
	if !ok {
		return
	}
	defer h.cleanupExchanges(c.Update.Message.ForwardFromChat.ID, c.Update.Message.ForwardFromMessageID)

	tweetIDFilesPostingProcessing, ok := h.Exchange.Load(baseKey + "/processing")
	if ok && tweetIDFilesPostingProcessing == true {
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
		"tweet_id":                tweetID,
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

	mediasRaw, ok := h.Exchange.Load(baseKey + "/medias")
	if !ok {
		h.Logger.WithFields(loggerFields).Error("medias not found")
		return
	}

	medias, ok := mediasRaw.([]*FetchedTweetMedia)
	if !ok {
		h.Logger.WithFields(loggerFields).Error("medias not found, incorrect type, type is not []*FetchedTweetMedia")
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
	thumbnailImages := make([]*bytes.Buffer, len(medias))
	for i, media := range medias {
		if media.Type != twitter_public_types.TweetLegacyExtendedEntityMediaTypePhoto {
			continue
		}

		img, err := imaging.Decode(bytes.NewReader(media.Body.Bytes()))
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

	h.Logger.Info("sending medias to discussion group...")
	mediaGroupConfig := tgbotapi.MediaGroupConfig{
		ReplyToMessageID: c.Update.Message.MessageID,
		ChatID:           c.Update.Message.Chat.ID,
		Media:            make([]interface{}, 0, len(medias)),
	}

	for i, media := range medias {
		parsedURL, err := url.Parse(media.URL)
		if err != nil {
			continue
		}

		parsedURL.RawQuery = ""
		parsedURL.Fragment = ""

		file := tgbotapi.FileBytes{
			Name: fmt.Sprintf("twitter-by-%s-%s-%d%s",
				authorName,
				tweetID,
				i,
				filepath.Ext(parsedURL.String()),
			),
			Bytes: media.OriginalBody.Bytes(),
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

	h.Logger.WithFields(loggerFields).Infof("%d images sent as comment of channel post in discussion group", len(medias))

	for _, i := range medias {
		i.Body.Reset()
		i.OriginalBody.Reset()
	}
}

package discord

import (
	"context"

	"github.com/ayumi-otosaka-314/brec-pp/bilibili"
)

type updateMessage struct {
	roomID          uint64
	messageID       string
	message         *WebhookMessage
	usingEmbedImage embedImageMapper
}

type embedImageMapper func(info *bilibili.LiveInfo) *MessageEmbedImage

func usingCover(info *bilibili.LiveInfo) *MessageEmbedImage {
	return &MessageEmbedImage{URL: info.Data.Room.Cover}
}

func usingKeyframe(info *bilibili.LiveInfo) *MessageEmbedImage {
	return &MessageEmbedImage{URL: info.Data.Room.Keyframe}
}

func (n *notifier) updateImages(ctx context.Context, updateMsg *updateMessage) error {
	message := updateMsg.message
	if len(message.Embeds) < 1 {
		n.logger.Warn("message does not have embed content to be updated")
		return nil
	}

	liveInfo, err := n.biliClient.GetLiveInfo(ctx, updateMsg.roomID)
	if err != nil {
		return err
	}

	message.Embeds[0].Thumbnail = &MessageEmbedThumbnail{URL: liveInfo.Data.Streamer.Base.Avatar}
	if updateMsg.usingEmbedImage != nil {
		message.Embeds[0].Image = updateMsg.usingEmbedImage(liveInfo)
	}

	return n.client.Edit(ctx, updateMsg.messageID, message)
}

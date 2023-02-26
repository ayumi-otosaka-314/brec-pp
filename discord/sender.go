package discord

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ayumi-otosaka-314/brec-pp/bilibili"
	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/notification"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
)

func NewNotifier(
	logger *zap.Logger,
	webhookURL string,
	storageSvc storage.Service,
) notification.Service {
	updateQueue := make(chan *updateMessage, 32)

	n := &notifier{
		logger:      logger,
		storageSvc:  storageSvc,
		client:      NewClient(logger, webhookURL),
		updateQueue: updateQueue,
	}

	go func() {
		for updateMsg := range updateQueue {
			ctx, _ := context.WithTimeout(context.Background(), 45&time.Second)
			if err := n.updateImages(ctx, updateMsg); err != nil {
				n.logger.Error("error updating image async", zap.Error(err))
			}
		}
	}()

	return n
}

type notifier struct {
	logger      *zap.Logger
	storageSvc  storage.Service
	client      Client
	updateQueue chan *updateMessage
	biliClient  bilibili.Client
}

func (n *notifier) OnRecordStart(
	ctx context.Context,
	eventTime time.Time,
	eventData *brec.EventDataSession,
) error {
	message := &WebhookMessage{
		Embeds: []*MessageEmbed{{
			Author: &MessageEmbedAuthor{
				Name: eventData.StreamerName,
				URL:  fmt.Sprintf("https://live.bilibili.com/%d", eventData.RoomID),
			},
			Title:       "Recording started",
			Type:        WebhookEmbedType,
			Description: eventData.Title,
			Timestamp:   eventTime.Format(time.RFC3339),
			Color:       0x0099FF,
			Fields: []*MessageEmbedField{{
				Name:  "Available Space on Recoder",
				Value: n.safeGetAvailableCapacity(),
			}},
		}},
	}

	response, err := n.client.Send(ctx, message)
	if err != nil {
		return errors.Wrap(err, "error sending OnRecordStart notification to discord")
	}

	select {
	case n.updateQueue <- &updateMessage{
		roomID:          eventData.RoomID,
		messageID:       response.ID,
		message:         message,
		usingEmbedImage: usingCover,
	}:
	case <-ctx.Done():
		n.logger.Error("unable to send message for update", zap.Error(ctx.Err()))
	}

	return nil
}

func (n *notifier) safeGetAvailableCapacity() string {
	availSpace, err := n.storageSvc.GetAvailableCapacity()
	if err != nil {
		n.logger.Error("error getting available capacity", zap.Error(err))
		return "error"
	}
	return fmt.Sprintf("%.3f GB", float64(availSpace)/storage.GigaBytes)
}

func (n *notifier) OnRecordReady(
	ctx context.Context,
	eventTime time.Time,
	eventData *brec.EventDataFileClose,
) error {
	message := &WebhookMessage{
		Embeds: []*MessageEmbed{{
			Author: &MessageEmbedAuthor{
				Name: eventData.StreamerName,
				URL:  fmt.Sprintf("https://live.bilibili.com/%d", eventData.RoomID),
			},
			Title: "Recording file ready for upload",
			Type:  WebhookEmbedType,
			Description: fmt.Sprintf(
				"Recording file of livestream [%s] is ready\nUploading now...",
				eventData.Title,
			),
			Timestamp: eventTime.Format(time.RFC3339),
			Color:     0x00FF99,
			Fields: []*MessageEmbedField{
				{
					Name:  "File Name",
					Value: path.Base(eventData.RelativePath),
				},
				{
					Name:   "File Size",
					Value:  fmt.Sprintf("%.3f GB", float64(eventData.FileSize)/storage.GigaBytes),
					Inline: true,
				},
				{
					Name:   "Recording Duration",
					Value:  time.Duration(eventData.Duration * float64(time.Second)).String(),
					Inline: true,
				},
			},
		}},
	}

	response, err := n.client.Send(ctx, message)
	if err != nil {
		return errors.Wrap(err, "error sending OnRecordReady notification to discord")
	}

	select {
	case n.updateQueue <- &updateMessage{
		roomID:          eventData.RoomID,
		messageID:       response.ID,
		message:         message,
		usingEmbedImage: usingKeyframe,
	}:
	case <-ctx.Done():
		n.logger.Error("unable to send message for update", zap.Error(ctx.Err()))
	}

	return nil
}

func (n *notifier) OnUploadComplete(
	ctx context.Context,
	timestamp time.Time,
	eventData *brec.EventDataFileClose,
	uploadDuration time.Duration,
) error {
	message := &WebhookMessage{
		Embeds: []*MessageEmbed{{
			Author: &MessageEmbedAuthor{
				Name: eventData.StreamerName,
				URL:  fmt.Sprintf("https://live.bilibili.com/%d", eventData.RoomID),
			},
			Title:     "Upload completed",
			Type:      WebhookEmbedType,
			Timestamp: timestamp.Format(time.RFC3339),
			Color:     0x99FF00,
			Fields: []*MessageEmbedField{
				{
					Name:  "File Name",
					Value: path.Base(eventData.RelativePath),
				},
				{
					Name:  "Upload Duration",
					Value: uploadDuration.String(),
				},
			},
		}},
	}

	response, err := n.client.Send(ctx, message)
	if err != nil {
		return errors.Wrap(err, "error sending OnUploadComplete notification to discord")
	}

	select {
	case n.updateQueue <- &updateMessage{
		roomID:          eventData.RoomID,
		messageID:       response.ID,
		message:         message,
		usingEmbedImage: nil, // not using embed image
	}:
	case <-ctx.Done():
		n.logger.Error("unable to send message for update", zap.Error(ctx.Err()))
	}

	return nil
}

func (n *notifier) Alert(ctx context.Context, msg string, err error) {
	if _, sendErr := n.client.Send(
		ctx,
		&WebhookMessage{
			Embeds: []*MessageEmbed{{
				Title: fmt.Sprintf("[Alert]"),
				Type:  WebhookEmbedType,
				Description: strings.Join([]string{
					"error happened in brec-pp:",
					msg,
				}, "\n"),
				Timestamp: time.Now().Format(time.RFC3339),
				Color:     0xFF0099,
				Fields: []*MessageEmbedField{
					{
						Name:  "Error",
						Value: err.Error(),
					},
				},
			}},
		}); sendErr != nil {
		n.logger.Error("error send alert to discord", zap.Error(sendErr))
	}
}

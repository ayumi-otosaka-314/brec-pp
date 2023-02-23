package discord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/notification"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
	"github.com/ayumi-otosaka-314/brec-pp/streamer"
)

func NewNotifier(
	logger *zap.Logger,
	webhookURL string,
	storageSvc storage.Service,
) notification.Service {
	updateQueue := make(chan *updateMessage, 32)
	go func() {
		for updateMsg := range updateQueue {
			logger.Warn("message to be updated",
				zap.Uint64("roomID", updateMsg.roomID),
				zap.String("sessionID", updateMsg.sessionID),
				zap.String("messageID", updateMsg.messageID),
			)
		}
	}()

	webhook, err := url.Parse(webhookURL)
	if err != nil {
		panic(err)
	}
	query := webhook.Query()
	query.Set("wait", "true")
	webhook.RawQuery = query.Encode()

	return &notifier{
		logger:      logger,
		webhookURL:  webhook.String(),
		storageSvc:  storageSvc,
		client:      http.DefaultClient,
		updateQueue: updateQueue,
	}
}

type notifier struct {
	logger               *zap.Logger
	webhookURL           string
	storageSvc           storage.Service
	client               *http.Client
	updateQueue          chan *updateMessage
	streamerMetaRegistry streamer.MetaRegistry
}

const (
	pocAvatar = "https://i2.hdslb.com/bfs/face/75ccf0dfbf9a4e56ee8d62115465f467f7e953aa.jpg"
)

func (n *notifier) OnRecordStart(
	ctx context.Context,
	eventTime time.Time,
	eventData *brec.EventDataSession,
) error {
	message := &webhookMessage{
		Embeds: []*messageEmbed{{
			Author: &messageEmbedAuthor{
				Name: eventData.StreamerName,
				URL:  fmt.Sprintf("https://live.bilibili.com/%d", eventData.RoomID),
			},
			Title:       "Recording started",
			Type:        webhookEmbedType,
			Description: eventData.Title,
			Thumbnail:   &messageEmbedThumbnail{URL: pocAvatar},
			Timestamp:   eventTime.Format(time.RFC3339),
			Color:       0x0099FF,
			Fields: []*messageEmbedField{{
				Name:  "Available Space on Recoder",
				Value: n.safeGetAvailableCapacity(),
			}},
		}},
	}
	return n.onMessage(ctx, message, eventData.RoomID, eventData.SessionID)
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
	message := &webhookMessage{
		Embeds: []*messageEmbed{{
			Author: &messageEmbedAuthor{
				Name: eventData.StreamerName,
				URL:  fmt.Sprintf("https://live.bilibili.com/%d", eventData.RoomID),
			},
			Title: "Recording file ready for upload",
			Type:  webhookEmbedType,
			Description: fmt.Sprintf(
				"Recording file of livestream [%s] is ready\nUploading now...",
				eventData.Title,
			),
			Thumbnail: &messageEmbedThumbnail{URL: pocAvatar},
			Timestamp: eventTime.Format(time.RFC3339),
			Color:     0x00FF99,
			Fields: []*messageEmbedField{
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
	return n.onMessage(ctx, message, eventData.RoomID, eventData.SessionID)
}

func (n *notifier) OnUploadComplete(
	ctx context.Context,
	timestamp time.Time,
	eventData *brec.EventDataFileClose,
	uploadDuration time.Duration,
) error {
	message := &webhookMessage{
		Embeds: []*messageEmbed{{
			Author: &messageEmbedAuthor{
				Name: eventData.StreamerName,
				URL:  fmt.Sprintf("https://live.bilibili.com/%d", eventData.RoomID),
			},
			Title:     "Upload completed",
			Type:      webhookEmbedType,
			Thumbnail: &messageEmbedThumbnail{URL: pocAvatar},
			Timestamp: timestamp.Format(time.RFC3339),
			Color:     0x99FF00,
			Fields: []*messageEmbedField{
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
	return n.onMessage(ctx, message, eventData.RoomID, eventData.SessionID)

}

func (n *notifier) Alert(msg string, err error) {
	if sendErr := n.onMessage(
		context.Background(),
		&webhookMessage{
			Embeds: []*messageEmbed{{
				Title: fmt.Sprintf("[Alert]"),
				Type:  webhookEmbedType,
				Description: strings.Join([]string{
					"error happened in brec-pp:",
					msg,
				}, "\n"),
				Timestamp: time.Now().Format(time.RFC3339),
				Color:     0xFF0099,
				Fields: []*messageEmbedField{
					{
						Name:  "Error",
						Value: err.Error(),
					},
				},
			}},
		},
		0, "",
	); sendErr != nil {
		n.logger.Error("error send alert to discord", zap.Error(sendErr))
	}
}

func (n *notifier) onMessage(
	ctx context.Context,
	message *webhookMessage,
	roomID uint64,
	sessionID string,
) error {
	raw, err := jsoniter.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error marshaling message")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		n.webhookURL,
		bytes.NewReader(raw),
	)
	if err != nil {
		return errors.Wrap(err, "error creating discord webhook request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending message to discord webhook")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		n.logger.Error("error sending message to discord webhook",
			zap.String("responseStatus", resp.Status),
			func() zapcore.Field {
				if rb, err := io.ReadAll(resp.Body); err != nil {
					return zap.Error(err)
				} else {
					return zap.String("response body", string(rb))
				}
			}(),
		)
		return errors.New("error sending message to discord webhook")
	}

	response := &webhookResponse{}
	if err = jsoniter.NewDecoder(resp.Body).Decode(response); err != nil {
		n.logger.Error("unable to decode discord webhook response body", zap.Error(err))
		return nil
	}

	select {
	case n.updateQueue <- &updateMessage{
		roomID:    roomID,
		sessionID: sessionID,
		messageID: response.ID,
		message:   message,
	}:
	case <-ctx.Done():
		n.logger.Error("error updating message", zap.Error(ctx.Err()))
	}
	return nil
}

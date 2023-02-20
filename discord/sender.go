package discord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
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
	"github.com/ayumi-otosaka-314/brec-pp/storage/localdrive"
)

func NewNotifier(
	logger *zap.Logger,
	webhookURL string,
	storageSvc *localdrive.Service,
) notification.Service {
	return &notifier{
		logger:     logger,
		webhookURL: webhookURL,
		storageSvc: storageSvc,
		client:     http.DefaultClient,
	}
}

type notifier struct {
	logger     *zap.Logger
	webhookURL string
	storageSvc *localdrive.Service
	client     *http.Client
}

func (n *notifier) OnRecordStart(
	ctx context.Context,
	eventTime time.Time,
	eventData *brec.EventDataSession,
) error {
	return n.onMessage(
		ctx,
		&webhookMessage{
			Embeds: []*webhookMessageEmbed{{
				Title: fmt.Sprintf("[%s] recording started", eventData.StreamerName),
				Type:  webhookEmbedType,
				Description: fmt.Sprintf(
					"Recording livestream [%s] from streamer [%s]",
					eventData.Title, eventData.StreamerName,
				),
				Timestamp: eventTime.Format(time.RFC3339),
				Color:     0x0099FF,
				Fields: []*webhookMessageEmbedField{{
					Name:  "Available Space on Recoder",
					Value: n.safeGetAvailableCapacity(),
				}},
			}},
		},
	)
}

func (n *notifier) safeGetAvailableCapacity() string {
	availSpace, err := n.storageSvc.GetAvailableCapacity()
	if err != nil {
		n.logger.Error("error getting available capacity", zap.Error(err))
		return "error"
	}
	return fmt.Sprintf("%.3f GB", float64(availSpace)/storage.GigaBytes)
}

func (n *notifier) OnRecordFinish(
	ctx context.Context,
	eventTime time.Time,
	eventData *brec.EventDataFileClose,
) error {
	return n.onMessage(
		ctx,
		&webhookMessage{
			Embeds: []*webhookMessageEmbed{{
				Title: fmt.Sprintf("[%s] recording file ready", eventData.StreamerName),
				Type:  webhookEmbedType,
				Description: fmt.Sprintf(
					"Recording file for livestream [%s] from streamer [%s] is ready\nUploading now...",
					eventData.Title, eventData.StreamerName,
				),
				Timestamp: eventTime.Format(time.RFC3339),
				Color:     0x00FF99,
				Fields: []*webhookMessageEmbedField{
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
		},
	)
}

func (n *notifier) OnUploadComplete(
	ctx context.Context,
	timestamp time.Time,
	uploadDuration time.Duration,
	streamerName string,
	fileName string,
) error {
	return n.onMessage(
		ctx,
		&webhookMessage{
			Embeds: []*webhookMessageEmbed{{
				Title: fmt.Sprintf("[%s] upload completed", streamerName),
				Type:  webhookEmbedType,
				Description: fmt.Sprintf(
					"Completed uploading livestream recording from streamer [%s]", streamerName,
				),
				Timestamp: timestamp.Format(time.RFC3339),
				Color:     0x99FF00,
				Fields: []*webhookMessageEmbedField{
					{
						Name:  "File Name",
						Value: path.Base(fileName),
					},
					{
						Name:  "Upload Duration",
						Value: uploadDuration.String(),
					},
				},
			}},
		},
	)
}

func (n *notifier) Alert(msg string, err error) {
	if err := n.onMessage(
		context.Background(),
		&webhookMessage{
			Embeds: []*webhookMessageEmbed{{
				Title: fmt.Sprintf("[Alert]"),
				Type:  webhookEmbedType,
				Description: strings.Join([]string{
					"error happened in brec-pp",
					msg,
				}, "\n"),
				Timestamp: time.Now().Format(time.RFC3339),
				Color:     0xFF0099,
				Fields: []*webhookMessageEmbedField{
					{
						Name:  "Error Message",
						Value: err.Error(),
					},
				},
			}},
		},
	); err != nil {
		n.logger.Error("error send alert to discord", zap.Error(err))
	}
}

func (n *notifier) onMessage(ctx context.Context, message *webhookMessage) error {
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

	response, err := n.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending message to discord webhook")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		n.logger.Error("error sending message to discord webhook",
			zap.String("responseStatus", response.Status),
			func() zapcore.Field {
				if rb, err := io.ReadAll(response.Body); err != nil {
					return zap.Error(err)
				} else {
					return zap.String("response body", string(rb))
				}
			}(),
		)
		return errors.New("error sending message to discord webhook")
	}
	return nil
}

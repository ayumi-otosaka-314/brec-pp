package discord

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Client interface {
	Send(context.Context, *WebhookMessage) (*WebhookResponse, error)
	Edit(context.Context, string, *WebhookMessage) error
}

func NewClient(
	logger *zap.Logger,
	webhookURL string,
) Client {
	waitedWebhook, err := url.Parse(webhookURL)
	if err != nil {
		panic(err)
	}
	query := waitedWebhook.Query()
	query.Set("wait", "true")
	waitedWebhook.RawQuery = query.Encode()

	return &client{
		logger:           logger,
		httpClient:       http.DefaultClient,
		webhookURL:       webhookURL,
		waitedWebhookURL: waitedWebhook.String(),
	}
}

type client struct {
	logger           *zap.Logger
	httpClient       *http.Client
	webhookURL       string
	waitedWebhookURL string
}

func (c *client) Send(
	ctx context.Context,
	message *WebhookMessage,
) (*WebhookResponse, error) {
	raw, err := jsoniter.Marshal(message)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling message")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.waitedWebhookURL,
		bytes.NewReader(raw),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating discord webhook request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error sending message to discord webhook")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		c.logger.Error("unexpected response status from discord webhook",
			zap.String("responseStatus", resp.Status),
			func() zapcore.Field {
				if rb, err := io.ReadAll(resp.Body); err != nil {
					return zap.Error(err)
				} else {
					return zap.String("responseBody", string(rb))
				}
			}(),
		)
		return nil, errors.New("unexpected response status from discord webhook")
	}

	response := &WebhookResponse{}
	if err = jsoniter.NewDecoder(resp.Body).Decode(response); err != nil {
		c.logger.Error("unable to decode discord webhook response body", zap.Error(err))
		return nil, errors.Wrap(err, "unable to decode discord webhook response body")
	}

	return response, nil
}

func (c *client) Edit(
	ctx context.Context,
	messageID string,
	message *WebhookMessage,
) error {
	raw, err := jsoniter.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error marshaling message")
	}

	editWebhookURL, err := url.JoinPath(c.webhookURL, "messages", messageID)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		editWebhookURL,
		bytes.NewReader(raw),
	)
	if err != nil {
		return errors.Wrap(err, "error creating discord webhook edit request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending edit request to discord webhook")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		c.logger.Error("unexpected response status from discord edit webhook",
			zap.String("responseStatus", resp.Status),
			func() zapcore.Field {
				if rb, err := io.ReadAll(resp.Body); err != nil {
					return zap.Error(err)
				} else {
					return zap.String("responseBody", string(rb))
				}
			}(),
		)
		return errors.New("unexpected response status from discord edit webhook")
	}

	return nil
}

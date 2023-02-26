package bilibili

import (
	"context"
	"fmt"
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Client interface {
	GetLiveInfo(ctx context.Context, roomID uint64) (*LiveInfo, error)
}

type client struct {
	logger     *zap.Logger
	httpClient *http.Client
}

func NewClient(logger *zap.Logger) Client {
	return &client{
		logger:     logger,
		httpClient: http.DefaultClient,
	}
}

func (c *client) GetLiveInfo(ctx context.Context, roomID uint64) (*LiveInfo, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%d", roomID),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating bilibili getLiveInfoByRoom request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error requesting bilibili getLiveInfoByRoom endpoint")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("unexpected response from bilibili getLiveInfoByRoom endpoint",
			zap.String("responseStatus", resp.Status),
			func() zapcore.Field {
				if rb, err := io.ReadAll(resp.Body); err != nil {
					return zap.Error(err)
				} else {
					return zap.String("responseBody", string(rb))
				}
			}(),
		)
		return nil, errors.New("unexpected response from bilibili getLiveInfoByRoom endpoint")
	}

	response := &LiveInfo{}
	if err = jsoniter.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, errors.Wrap(err, "unable to decode bilibili getLiveInfoByRoom response")
	}

	return response, nil
}

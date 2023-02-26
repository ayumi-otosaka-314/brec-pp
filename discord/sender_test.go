//go:build testRealAPI

package discord

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/ayumi-otosaka-314/brec-pp/bilibili"
	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
	"github.com/ayumi-otosaka-314/brec-pp/storage/localdrive"
)

const testWebhook = "REPLACE_WITH_TEST_WEBHOOK"

func Test_notifier_OnRecordStart(t *testing.T) {
	logger := zaptest.NewLogger(t)
	n := NewNotifier(
		logger,
		testWebhook,
		localdrive.New(logger, "."),
		&dummyBiliClient{logger},
	)

	assert.NoError(t, n.OnRecordStart(
		context.Background(),
		time.Now(),
		&brec.EventDataSession{
			EventDataBase: brec.EventDataBase{
				StreamerName: "テスト配信者",
				Title:        "テスト配信枠",
			},
		},
	))
}

func Test_notifier_OnRecordFinish(t *testing.T) {
	logger := zaptest.NewLogger(t)
	n := NewNotifier(
		logger,
		testWebhook,
		localdrive.New(logger, "."),
		&dummyBiliClient{logger},
	)

	assert.NoError(t, n.OnRecordReady(
		context.Background(),
		time.Now(),
		&brec.EventDataFileClose{
			EventDataBase: brec.EventDataBase{
				StreamerName: "テスト配信者",
				Title:        "テスト配信枠",
			},
			RelativePath: "/usr/local/var/brec/test-record.flv",
			FileSize: func() uint64 {
				size := 1.24 * storage.GigaBytes
				return uint64(size)
			}(),
			Duration: 4382.572,
		},
	))
}

func Test_notifier_OnUploadFinish(t *testing.T) {
	logger := zaptest.NewLogger(t)
	n := NewNotifier(
		logger,
		testWebhook,
		localdrive.New(logger, "."),
		&dummyBiliClient{logger},
	)

	assert.NoError(t, n.OnUploadComplete(
		context.Background(),
		time.Now(),
		&brec.EventDataFileClose{
			EventDataBase: brec.EventDataBase{
				StreamerName: "テスト配信者",
			},
			RelativePath: "/usr/local/var/brec/test-record.flv",
		},
		18*time.Minute,
	))
}

func Test_notifier_Alert(t *testing.T) {
	logger := zaptest.NewLogger(t)
	n := NewNotifier(
		logger,
		testWebhook,
		localdrive.New(logger, "."),
		&dummyBiliClient{logger},
	)

	n.Alert(context.Background(), "test message", errors.New("test error"))
}

type dummyBiliClient struct {
	logger *zap.Logger
}

func (d *dummyBiliClient) GetLiveInfo(ctx context.Context, roomID uint64) (*bilibili.LiveInfo, error) {
	d.logger.Info("bilibili.Client.GetLiveInfo called", zap.Uint64("roomID", roomID))
	return nil, errors.New("dummy client")
}

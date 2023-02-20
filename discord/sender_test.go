package discord

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
	"github.com/ayumi-otosaka-314/brec-pp/storage/localdrive"
)

const testWebhook = "REPLACE_WITH_TEST_WEBHOOK"

func Test_notifier_OnRecordStart(t *testing.T) {
	logger, _ := zap.NewProduction()

	n := &notifier{
		logger:     logger,
		webhookURL: testWebhook,
		storageSvc: localdrive.NewService("."),
		client:     http.DefaultClient,
	}

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
	logger, _ := zap.NewProduction()
	n := &notifier{
		logger:     logger,
		webhookURL: testWebhook,
		client:     http.DefaultClient,
	}

	assert.NoError(t, n.OnRecordFinish(
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
	logger, _ := zap.NewProduction()
	n := &notifier{
		logger:     logger,
		webhookURL: testWebhook,
		client:     http.DefaultClient,
	}

	assert.NoError(t, n.OnUploadComplete(
		context.Background(),
		time.Now(),
		18*time.Minute,
		"テスト配信者",
		"test-record.flv",
	))
}

func Test_notifier_Alert(t *testing.T) {
	logger, _ := zap.NewProduction()
	n := &notifier{
		logger:     logger,
		webhookURL: testWebhook,
		client:     http.DefaultClient,
	}

	n.Alert("test message", errors.New("test error"))
}

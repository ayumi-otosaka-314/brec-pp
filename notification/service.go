package notification

import (
	"context"
	"time"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
)

type Service interface {
	OnRecordStart(context.Context, time.Time, *brec.EventDataSession) error
	OnRecordFinish(context.Context, time.Time, *brec.EventDataFileClose) error
	OnUploadComplete(
		context context.Context,
		timestamp time.Time,
		uploadDuration time.Duration,
		streamerName string,
		fileName string,
	) error
	Alert(string, error)
}

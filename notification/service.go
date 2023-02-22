package notification

import (
	"context"
	"time"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
)

type Service interface {
	OnRecordStart(context.Context, time.Time, *brec.EventDataSession) error
	OnRecordReady(context.Context, time.Time, *brec.EventDataFileClose) error
	OnUploadComplete(context.Context, time.Time, *brec.EventDataFileClose, time.Duration) error
	Alert(string, error)
}

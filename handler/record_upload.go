package handler

import (
	"context"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/notification"
	"github.com/ayumi-otosaka-314/brec-pp/upload"
)

func NewNotifyRecordUploadHandler(
	logger *zap.Logger,
	timeout time.Duration,
	notifier notification.Service,
	uploader upload.Service,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rootCtx, _ := context.WithTimeout(context.Background(), timeout)

		if r.Method != http.MethodPost {
			logger.Warn("unexpected HTTP method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		event := &brec.Event{}
		if err := jsoniter.NewDecoder(r.Body).Decode(event); err != nil {
			logger.Error("error decoding request body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		eventTime, err := event.GetTimestamp()
		if err != nil {
			logger.Error("error parsing event timestamp",
				zap.Error(err), zap.String("EventTimestamp", event.TimeStamp))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch event.Type {
		case brec.EventTypeSessionStarted:
			eventData := &brec.EventDataSession{}
			if err = jsoniter.Unmarshal(event.Data, eventData); err != nil {
				logger.Warn("error unmarshalling event data", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = notifier.OnRecordStart(rootCtx, eventTime, eventData)
		case brec.EventTypeFileClosed:
			eventData := &brec.EventDataFileClose{}
			if err = jsoniter.Unmarshal(event.Data, eventData); err != nil {
				logger.Warn("error unmarshalling event data", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = notifier.OnRecordFinish(rootCtx, eventTime, eventData)

			select {
			case uploader.Receive() <- eventData:
			case <-rootCtx.Done():
				err = rootCtx.Err()
			}
		default:
			logger.Debug("received unqualified event", zap.Object("Event", event))
		}

		if err != nil {
			logger.Warn("error processing event", zap.Object("Event", event), zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}
}

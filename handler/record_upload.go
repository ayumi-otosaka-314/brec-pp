package handler

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/storage/localdrive"
	"github.com/ayumi-otosaka-314/brec-pp/streamer"
)

func NewNotifyRecordUploadHandler(
	logger *zap.Logger,
	timeout time.Duration,
	localStorage localdrive.Service,
	streamerServiceRegistry streamer.ServiceRegistry,
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

			err = streamerServiceRegistry.
				GetNotifier(eventData.RoomID).
				OnRecordStart(rootCtx, eventTime, eventData)
		case brec.EventTypeFileOpening:
			eventData := &brec.EventDataFileOpen{}
			if err = jsoniter.Unmarshal(event.Data, eventData); err != nil {
				logger.Warn("error unmarshalling event data", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			traverseDepth := strings.Count(eventData.RelativePath, string(os.PathSeparator))
			err = localStorage.AsyncClean(rootCtx, traverseDepth)
		case brec.EventTypeFileClosed:
			eventData := &brec.EventDataFileClose{}
			if err = jsoniter.Unmarshal(event.Data, eventData); err != nil {
				logger.Warn("error unmarshalling event data", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err = streamerServiceRegistry.
				GetNotifier(eventData.RoomID).
				OnRecordReady(rootCtx, eventTime, eventData); err != nil {
				logger.Warn("error notifying on record finish; continue to upload", zap.Error(err))
			}

			select {
			case streamerServiceRegistry.GetUploader(eventData.RoomID).Receive() <- eventData:
				err = nil
			case <-rootCtx.Done():
				err = errors.Wrap(rootCtx.Err(), "unable to upload")
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

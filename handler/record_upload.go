package handler

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
	"github.com/ayumi-otosaka-314/brec-pp/storage/localdrive"
	"github.com/ayumi-otosaka-314/brec-pp/streamer"
)

func NewNotifyRecordUploadHandler(
	logger *zap.Logger,
	timeout time.Duration,
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

			go func(e *brec.EventDataFileOpen) {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
				if err := storage.EnsureCapacity(
					localdrive.WithTraverseDepth(ctx, strings.Count(e.RelativePath, string(os.PathSeparator))),
					uint64(5*storage.GigaBytes),
					streamerServiceRegistry.GetLocalStorage(eventData.RoomID),
				); err != nil {
					logger.Error("error cleaning local storage", zap.Error(err))
					streamerServiceRegistry.GetNotifier(eventData.RoomID).
						Alert("error cleaning local storage", err)
				}
			}(eventData)
		case brec.EventTypeFileClosed:
			eventData := &brec.EventDataFileClose{}
			if err = jsoniter.Unmarshal(event.Data, eventData); err != nil {
				logger.Warn("error unmarshalling event data", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			
			if err = streamerServiceRegistry.
				GetNotifier(eventData.RoomID).
				OnRecordFinish(rootCtx, eventTime, eventData); err != nil {
				logger.Warn("error notifying on record finish; continue to upload", zap.Error(err))
			}

			select {
			case streamerServiceRegistry.GetUploader(eventData.RoomID).Receive() <- eventData:
				err = nil
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

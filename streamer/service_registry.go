package streamer

import (
	"github.com/ayumi-otosaka-314/brec-pp/notification"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
	"github.com/ayumi-otosaka-314/brec-pp/upload"
)

type ServiceRegistry interface {
	GetNotifier(roomID uint64) notification.Service
	GetLocalStorage(roomID uint64) storage.Cleaner
	GetUploader(roomID uint64) upload.Service
}

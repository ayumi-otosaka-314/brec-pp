package registry

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ayumi-otosaka-314/brec-pp/config"
	"github.com/ayumi-otosaka-314/brec-pp/discord"
	"github.com/ayumi-otosaka-314/brec-pp/handler"
	"github.com/ayumi-otosaka-314/brec-pp/notification"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
	"github.com/ayumi-otosaka-314/brec-pp/storage/gdrive"
	"github.com/ayumi-otosaka-314/brec-pp/storage/localdrive"
	"github.com/ayumi-otosaka-314/brec-pp/streamer"
	"github.com/ayumi-otosaka-314/brec-pp/upload"
)

type Registry struct {
	conf   *config.Root
	logger *zap.Logger
}

func New(conf *config.Root) *Registry {
	return &Registry{
		conf:   conf,
		logger: NewLogger(),
	}
}

func NewLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

func (r *Registry) NewServer() *handler.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		r.conf.Server.Paths.RecordUpload,
		handler.NewNotifyRecordUploadHandler(
			r.logger,
			r.conf.Server.Timeout,
			r.NewServiceRegistry(),
		),
	)
	return handler.NewServer(r.logger, r.conf.Server.ListenAddress, mux)
}

func (r *Registry) CleanUp() {
	r.logger.Sync()
}

func (r *Registry) NewServiceRegistry() streamer.ServiceRegistry {
	mapping := make(map[uint64]*serviceEntry, len(r.conf.Services.Streamers))
	for _, entry := range r.conf.Services.Streamers {
		mapping[entry.RoomID] = r.newServiceEntry(entry.ServiceEntry)
	}
	return &serviceRegistry{
		mapping:      mapping,
		defaultEntry: r.newServiceEntry(r.conf.Services.Default),
	}
}

type serviceRegistry struct {
	mapping      map[uint64]*serviceEntry
	defaultEntry *serviceEntry
}

type serviceEntry struct {
	notifier     notification.Service
	localStorage storage.Cleaner
	uploader     upload.Service
}

func (r *Registry) newServiceEntry(conf config.ServiceEntry) *serviceEntry {
	localStorage := localdrive.New(r.logger, conf.Storage.RootPath)
	notifier := discord.NewNotifier(
		r.logger,
		conf.Discord.WebhookURL,
		localStorage,
	)
	return &serviceEntry{
		notifier:     notifier,
		localStorage: localStorage,
		uploader: gdrive.NewUploadService(
			r.logger,
			&conf.Storage.GoogleDrive,
			conf.Storage.RootPath,
			notifier,
		),
	}
}

func (s *serviceRegistry) GetNotifier(roomID uint64) notification.Service {
	return s.getServiceEntry(roomID).notifier
}

func (s *serviceRegistry) GetLocalStorage(roomID uint64) storage.Cleaner {
	return s.getServiceEntry(roomID).localStorage
}

func (s *serviceRegistry) GetUploader(roomID uint64) upload.Service {
	return s.getServiceEntry(roomID).uploader
}

func (s *serviceRegistry) getServiceEntry(roomID uint64) *serviceEntry {
	if entry, ok := s.mapping[roomID]; ok {
		return entry
	}
	return s.defaultEntry
}

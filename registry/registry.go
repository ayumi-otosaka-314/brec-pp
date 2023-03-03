package registry

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ayumi-otosaka-314/brec-pp/bilibili"
	"github.com/ayumi-otosaka-314/brec-pp/config"
	"github.com/ayumi-otosaka-314/brec-pp/discord"
	"github.com/ayumi-otosaka-314/brec-pp/handler"
	"github.com/ayumi-otosaka-314/brec-pp/notification"
	"github.com/ayumi-otosaka-314/brec-pp/storage/gdrive"
	"github.com/ayumi-otosaka-314/brec-pp/storage/localdrive"
	"github.com/ayumi-otosaka-314/brec-pp/streamer"
	"github.com/ayumi-otosaka-314/brec-pp/upload"
)

type Registry struct {
	conf         *config.Root
	logger       *zap.Logger
	localStorage localdrive.Service
}

func New(conf *config.Root) *Registry {
	logger := NewLogger()
	return &Registry{
		conf:         conf,
		logger:       logger,
		localStorage: NewLocalStorage(logger, &conf.LocalStorage),
	}
}

func NewLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := cfg.Build(
		zap.AddStacktrace(zap.WarnLevel),
		zap.AddCaller(),
	)
	if err != nil {
		panic(err)
	}
	return logger
}

func NewLocalStorage(logger *zap.Logger, conf *config.LocalStorage) localdrive.Service {
	return localdrive.New(
		logger,
		conf.RootPath,
		conf.CleanInterval,
		conf.ReservedCapacity,
	)
}

func (r *Registry) NewServer() *handler.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(
		r.conf.Server.Paths.RecordUpload,
		handler.NewNotifyRecordUploadHandler(
			r.logger,
			r.conf.Server.Timeout,
			r.localStorage,
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
		mapping[entry.RoomID] = r.newServiceEntry(entry.ServiceEntry, r.conf.LocalStorage.RootPath)
	}
	return &serviceRegistry{
		mapping:      mapping,
		defaultEntry: r.newServiceEntry(r.conf.Services.Default, r.conf.LocalStorage.RootPath),
	}
}

func (r *Registry) newBiliClient() bilibili.Client {
	return bilibili.NewClient(r.logger)
}

type serviceRegistry struct {
	mapping      map[uint64]*serviceEntry
	defaultEntry *serviceEntry
}

type serviceEntry struct {
	notifier notification.Service
	uploader upload.Service
}

func (r *Registry) newServiceEntry(conf config.ServiceEntry, localRootPath string) *serviceEntry {
	notifier := discord.NewNotifier(
		r.logger,
		conf.Notification.Discord.WebhookURL,
		r.localStorage,
		r.newBiliClient(),
	)
	return &serviceEntry{
		notifier: notifier,
		uploader: gdrive.NewUploadService(
			r.logger,
			&conf.Upload.GoogleDrive,
			localRootPath,
			notifier,
		),
	}
}

func (s *serviceRegistry) GetNotifier(roomID uint64) notification.Service {
	return s.getServiceEntry(roomID).notifier
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

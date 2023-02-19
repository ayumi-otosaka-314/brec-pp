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
			r.NewNotifier(),
			r.NewUploadService(),
		),
	)
	return handler.NewServer(r.logger, r.conf.Server.ListenAddress, mux)
}

func (r *Registry) CleanUp() {
	r.logger.Sync()
}

func (r *Registry) NewNotifier() notification.Service {
	return discord.NewNotifier(
		r.logger,
		r.conf.Discord.WebhookURL,
		r.NewStorageService(),
	)
}

func (r *Registry) NewStorageService() *storage.Service {
	return storage.NewService(r.conf.Storage.RootPath)
}

func (r *Registry) NewUploadService() upload.Service {
	return gdrive.NewUploadService(
		r.logger,
		&r.conf.Storage.GoogleDrive,
		r.conf.Storage.RootPath,
		r.NewNotifier(),
	)
}

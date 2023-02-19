package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/ayumi-otosaka-314/brec-pp/brec"
	"github.com/ayumi-otosaka-314/brec-pp/config"
	"github.com/ayumi-otosaka-314/brec-pp/notification"
	"github.com/ayumi-otosaka-314/brec-pp/storage"
	"github.com/ayumi-otosaka-314/brec-pp/upload"
)

type service struct {
	logger               *zap.Logger
	config               *jwt.Config
	timeout              time.Duration
	reservedCapacity     uint64
	gdriveParentFolderID string
	localRootPath        string
	notifier             notification.Service
	receive              chan *brec.EventDataFileClose
}

func NewUploadService(
	logger *zap.Logger,
	gdriveConfig *config.GoogleDrive,
	localRootPath string,
	notifier notification.Service,
) upload.Service {
	conf, err := fromServiceAccount(gdriveConfig.CredentialPath)
	if err != nil {
		panic(err)
	}
	svc := &service{
		logger:               logger,
		config:               conf,
		timeout:              gdriveConfig.Timeout,
		reservedCapacity:     gdriveConfig.ReservedCapacity,
		gdriveParentFolderID: gdriveConfig.ParentFolderID,
		localRootPath:        localRootPath,
		notifier:             notifier,
		receive:              make(chan *brec.EventDataFileClose, 16),
	}
	go svc.doReceive()
	return svc
}

func fromServiceAccount(credentialPath string) (*jwt.Config, error) {
	b, err := os.ReadFile(credentialPath)
	if err != nil {
		return nil, errors.Wrap(err, "error reading credential file")
	}
	var c = struct {
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
		ClientEmail  string `json:"client_email"`
		TokenURI     string `json:"token_uri"`
	}{}
	if err = json.Unmarshal(b, &c); err != nil {
		return nil, errors.Wrap(err, "err unmarshalling credential file")
	}
	return &jwt.Config{
		PrivateKeyID: c.PrivateKeyID,
		PrivateKey:   []byte(c.PrivateKey),
		Email:        c.ClientEmail,
		Scopes:       []string{drive.DriveScope},
		TokenURL:     c.TokenURI,
	}, nil
	//client := config.Client(context.Background())
	//return client
}

func (s *service) Receive() chan<- *brec.EventDataFileClose {
	return s.receive
}

func (s *service) doReceive() {
	for eventData := range s.receive {
		go func(streamerName, relativePath string) {
			if err := s.doUpload(streamerName, relativePath); err != nil {
				s.logger.Error("error uploading file",
					zap.String("streamerName", streamerName), zap.String("filePath", relativePath),
					zap.Error(err))
				s.notifier.Alert(fmt.Sprintf(
					"error uploading file [%s] for streamer [%s]",
					relativePath, streamerName,
				), err)
			}
		}(eventData.StreamerName, eventData.RelativePath)
	}
	s.logger.Warn("receive channel closed for google drive uploader")
}

func (s *service) doUpload(streamerName, relativePath string) error {
	start := time.Now()

	ctx, _ := context.WithTimeout(context.Background(), s.timeout)
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(s.config.Client(ctx)))
	if err != nil {
		return errors.Wrap(err, "unable to create google drive service")
	}

	if err = storage.EnsureCapacity(
		s.reservedCapacity,
		&cleaner{logger: s.logger, driveService: driveService},
	); err != nil {
		return errors.Wrap(err, "unable to ensure capacity")
	}

	uploadFile, err := os.Open(path.Join(s.localRootPath, relativePath))
	if err != nil {
		return errors.Wrap(err, "unable to open uploadFile file")
	}
	fileName := path.Base(relativePath)
	if _, err = driveService.Files.
		Create(&drive.File{
			Name:    fileName,
			Parents: []string{s.gdriveParentFolderID}},
		).
		Media(uploadFile).
		ProgressUpdater(s.logUploadProgress(fileName)).
		Do(); err != nil {
		return errors.Wrap(err, "unable to uploadFile")
	}

	return s.notifier.OnUploadComplete(ctx, time.Now(), time.Since(start), streamerName, fileName)
}

func (s *service) logUploadProgress(fileName string) googleapi.ProgressUpdater {
	return func(current, total int64) {
		s.logger.Debug(
			"upload progress update",
			zap.String("fileName", fileName),
			zap.Int64("current", current),
			zap.Int64("total", total),
		)
	}
}

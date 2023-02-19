package gdrive

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/drive/v3"

	"github.com/ayumi-otosaka-314/brec-pp/storage"
)

type cleaner struct {
	logger       *zap.Logger
	driveService *drive.Service
}

func (c *cleaner) GetAvailableCapacity() (uint64, error) {
	about, err := c.driveService.About.Get().Fields("storageQuota").Do()
	if err != nil {
		return 0, errors.Wrap(err, "unable to get gdrive usage")
	}
	return uint64(about.StorageQuota.Limit - about.StorageQuota.Usage), nil
}

func (c *cleaner) GetRemovables(ctx context.Context) (<-chan storage.Removable, error) {

	r, err := c.driveService.Files.
		List().
		OrderBy("modifiedTime").
		PageSize(5).
		Fields("nextPageToken, files(id, name, size)").
		Do()
	if err != nil {
		return nil, err
	}
	result := make(chan storage.Removable)
	go func() {
		defer close(result)

		for _, file := range r.Files {
			select {
			case result <- &removable{
				driveService: c.driveService,
				logger:       c.logger,
				fileID:       file.Id,
				name:         file.Name,
				size:         uint64(file.Size),
			}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return result, nil
}

type removable struct {
	driveService *drive.Service
	logger       *zap.Logger
	fileID       string
	name         string
	size         uint64
}

func (r *removable) Remove() error {
	r.logger.Debug("deleting file from google drive",
		zap.String("fileID", r.fileID), zap.String("name", r.name))
	return r.driveService.Files.Delete(r.fileID).Do()
}

func (r *removable) OccupiedSize() uint64 {
	return r.size
}

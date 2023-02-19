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
			removeResult := make(chan error)
			select {
			case result <- &removable{ // send result to downstream first, then perform action
				removeResult: removeResult,
				size:         uint64(file.Size),
			}:
				c.logger.Debug("deleting file from google drive",
					zap.String("fileID", file.Id), zap.String("name", file.Name))
				removeResult <- c.driveService.Files.Delete(file.Id).Do()
			case <-ctx.Done():
				return
			}
		}
	}()
	return result, nil
}

type removable struct {
	removeResult <-chan error
	size         uint64
}

func (r *removable) RemoveResult() <-chan error {
	return r.removeResult
}

func (r *removable) OccupiedSize() uint64 {
	return r.size
}

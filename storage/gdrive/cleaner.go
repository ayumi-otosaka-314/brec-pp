package gdrive

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/drive/v3"

	"github.com/ayumi-otosaka-314/brec-pp/storage"
)

// cleaner implements storage.Cleaner.
// It is used to clear old recordings on Google Drive to ensure capacity before uploading.
type cleaner struct {
	logger         *zap.Logger
	driveService   *drive.Service
	parentFolderID string
}

func (c *cleaner) GetAvailableCapacity() (uint64, error) {
	about, err := c.driveService.About.Get().Fields("storageQuota").Do()
	if err != nil {
		return 0, errors.Wrap(err, "unable to get gdrive usage")
	}
	return uint64(about.StorageQuota.Limit - about.StorageQuota.Usage), nil
}

func (c *cleaner) GetRemovables(ctx context.Context) (<-chan storage.DoRemove, error) {

	r, err := c.driveService.Files.
		List().
		Q(fmt.Sprintf(
			"mimeType != 'application/vnd.google-apps.folder' and '%s' in parents",
			c.parentFolderID,
		)).
		OrderBy("modifiedTime").
		PageSize(10).
		Fields("files(id, name, size)").
		Do()
	if err != nil {
		return nil, err
	}

	result := make(chan storage.DoRemove)
	go func() {
		defer close(result)

		for _, file := range r.Files {
			doRemove := func() (uint64, error) {
				c.logger.Debug("deleting file from google drive",
					zap.String("name", file.Name), zap.String("fileID", file.Id))
				return uint64(file.Size), c.driveService.Files.Delete(file.Id).Do()
			}
			select {
			case result <- doRemove:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
	return result, nil
}

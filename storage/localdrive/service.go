package localdrive

import (
	"context"

	"golang.org/x/sys/unix"

	"github.com/ayumi-otosaka-314/brec-pp/storage"
)

type service struct {
	rootPath string
}

func New(rootPath string) storage.Cleaner {
	return &service{rootPath: rootPath}
}

func (s *service) GetAvailableCapacity() (uint64, error) {
	var statfs unix.Statfs_t
	if err := unix.Statfs(s.rootPath, &statfs); err != nil {
		return 0, err
	}
	return statfs.Bavail * uint64(statfs.Bsize), nil
}

func (s *service) GetRemovables(ctx context.Context) (<-chan storage.DoRemove, error) {
	//TODO implement me
	panic("implement me")
}

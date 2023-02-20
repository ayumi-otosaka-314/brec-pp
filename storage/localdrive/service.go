package localdrive

import (
	"context"

	"golang.org/x/sys/unix"

	"github.com/ayumi-otosaka-314/brec-pp/storage"
)

type Service struct {
	rootPath string
}

func NewService(rootPath string) *Service {
	return &Service{rootPath: rootPath}
}

func (s *Service) GetAvailableCapacity() (uint64, error) {
	var statfs unix.Statfs_t
	if err := unix.Statfs(s.rootPath, &statfs); err != nil {
		return 0, err
	}
	return statfs.Bavail * uint64(statfs.Bsize), nil
}

func (s *Service) GetRemovables(ctx context.Context) (<-chan storage.DoRemove, error) {
	//TODO implement me
	panic("implement me")
}

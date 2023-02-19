package storage

import (
	"golang.org/x/sys/unix"
)

type Service struct {
	rootPath string
}

func NewService(rootPath string) *Service {
	return &Service{rootPath: rootPath}
}

const (
	KiloBytes float64 = 1024
	MegaBytes         = 1024 * KiloBytes
	GigaBytes         = 1024 * MegaBytes
)

func (s *Service) GetAvailableSpace() (uint64, error) {
	var statfs unix.Statfs_t
	if err := unix.Statfs(s.rootPath, &statfs); err != nil {
		return 0, err
	}
	return statfs.Bavail * uint64(statfs.Bsize), nil
}

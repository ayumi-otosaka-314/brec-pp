package localdrive

import (
	"context"
	"os"
	"path"
	"sort"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/ayumi-otosaka-314/brec-pp/storage"
)

type service struct {
	rootPath string
	logger   *zap.Logger
}

func New(logger *zap.Logger, rootPath string) storage.Cleaner {
	return &service{rootPath: rootPath, logger: logger}
}

func (s *service) GetAvailableCapacity() (uint64, error) {
	var statfs unix.Statfs_t
	if err := unix.Statfs(s.rootPath, &statfs); err != nil {
		return 0, err
	}
	return statfs.Bavail * uint64(statfs.Bsize), nil
}

func (s *service) GetRemovables(ctx context.Context) (<-chan storage.DoRemove, error) {
	var traverseDepth = 2 // traverse 2 levels by default.
	val := ctx.Value(keyTraverseDepth)
	if ctxDepth, ok := val.(int); ok && ctxDepth > 0 {
		traverseDepth = ctxDepth
	}

	entries := make([]*fileEntry, 0)
	if err := s.traverse(s.rootPath, traverseDepth, &entries); err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].lastModified.Before(entries[j].lastModified)
	})

	result := make(chan storage.DoRemove)
	go func() {
		defer close(result)

		for _, entry := range entries {
			doRemove := func() (uint64, error) {
				removePath := path.Join(entry.parentPath, entry.name)
				s.logger.Debug("deleting file from local drive", zap.String("path", removePath))
				return entry.size, os.Remove(path.Join(entry.parentPath, entry.name))
			}
			select {
			case result <- doRemove:
				continue
			case <-ctx.Done():
				s.logger.Debug("local drive get removable finished", zap.Error(ctx.Err()))
				return
			}
		}
	}()

	return result, nil
}

type fileEntry struct {
	size         uint64
	lastModified time.Time
	parentPath   string
	name         string
}

func (s *service) traverse(root string, depth int, result *[]*fileEntry) error {
	if depth < 0 {
		return nil
	}

	files, err := os.ReadDir(root)
	if err != nil {
		return errors.Wrap(err, "unable to read dir")
	}

	for _, file := range files {
		if file.IsDir() {
			if err = s.traverse(path.Join(root, file.Name()), depth-1, result); err != nil {
				return err
			}
		} else {
			info, err := file.Info()
			if err != nil {
				s.logger.Error("error getting file info",
					zap.String("parentPath", root), zap.String("fileName", file.Name()))
				return errors.Wrap(err, "error getting file info")
			}
			*result = append(*result, &fileEntry{
				size:         uint64(info.Size()),
				lastModified: info.ModTime(),
				parentPath:   root,
				name:         file.Name(),
			})
		}
	}

	return nil
}

type contextKey uint8

const (
	keyTraverseDepth contextKey = 1
)

func WithTraverseDepth(ctx context.Context, depth int) context.Context {
	return context.WithValue(ctx, keyTraverseDepth, depth)
}

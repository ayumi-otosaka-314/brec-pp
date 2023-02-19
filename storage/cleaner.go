package storage

import (
	"context"

	"github.com/pkg/errors"
)

type Cleaner interface {
	GetAvailableCapacity() (uint64, error)
	GetRemovables(context.Context) (<-chan Removable, error)
}

type Removable interface {
	Remove() error
	OccupiedSize() uint64
}

func EnsureCapacity(targetCapacity uint64, cleaner Cleaner) error {
	for {
		availCapacity, err := cleaner.GetAvailableCapacity()
		if err != nil {
			return errors.Wrap(err, "unable to check available bytes")
		}

		if availCapacity >= targetCapacity {
			return nil
		}

		if err = doEnsureCapacity(targetCapacity-availCapacity, cleaner); err != nil {
			return err
		}
	}
}

func doEnsureCapacity(cleanTarget uint64, cleaner Cleaner) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	removables, err := cleaner.GetRemovables(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get removables")
	}

	for removable := range removables {
		if err = removable.Remove(); err != nil {
			return errors.Wrap(err, "error removing object; stopping")
		}

		cleanTarget -= removable.OccupiedSize()
		if cleanTarget <= 0 {
			return nil
		}
	}

	return errors.New("no removable left")
}

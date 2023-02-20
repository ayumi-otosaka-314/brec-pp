package storage

import (
	"context"

	"github.com/pkg/errors"
)

type Cleaner interface {
	GetAvailableCapacity() (uint64, error)
	GetRemovables(context.Context) (<-chan DoRemove, error)
}

// DoRemove is the action to actually remove removable.
// It would return the space cleared in byte count, and error if any during cleaning.
type DoRemove func() (uint64, error)

func EnsureCapacity(targetCapacity uint64, cleaner Cleaner) error {
	const allowedIterations = 5
	for i := 0; i < allowedIterations; i++ {
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
	return errors.Errorf("unable to ensure capacity in [%d] iterations", allowedIterations)
}

func doEnsureCapacity(cleanTarget uint64, cleaner Cleaner) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	removables, err := cleaner.GetRemovables(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get removables")
	}

	for remove := range removables {
		clearedSize, err := remove()
		if err != nil {
			return errors.Wrap(err, "error removing object; stopping")
		}

		cleanTarget -= clearedSize
		if cleanTarget <= 0 {
			return nil
		}
	}
	return nil
}

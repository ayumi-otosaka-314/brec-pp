package storage

import (
	"context"

	"github.com/pkg/errors"
)

type Service interface {
	GetAvailableCapacity() (uint64, error)
}

type Cleaner interface {
	Service
	GetRemovables(context.Context) (<-chan DoRemove, error)
}

// DoRemove is the action to actually remove removable.
// It would return the space cleared in byte count, and error if any during cleaning.
type DoRemove func() (uint64, error)

func EnsureCapacity(ctx context.Context, targetCapacity uint64, cleaner Cleaner) error {
	const allowedIterations = 5
	for i := 0; i < allowedIterations; i++ {
		availCapacity, err := cleaner.GetAvailableCapacity()
		if err != nil {
			return errors.Wrap(err, "unable to check available bytes")
		}

		if availCapacity >= targetCapacity {
			return nil
		}

		if err = doEnsureCapacity(ctx, targetCapacity-availCapacity, cleaner); err != nil {
			return err
		}
	}
	return errors.Errorf("unable to ensure capacity in [%d] iterations", allowedIterations)
}

func doEnsureCapacity(parentCtx context.Context, cleanTarget uint64, cleaner Cleaner) error {
	ctx, cancel := context.WithCancel(parentCtx)
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

		// check before performing subtraction, to prevent overflow of uint64.
		if clearedSize >= cleanTarget {
			return nil
		}
		cleanTarget -= clearedSize
	}
	return nil
}

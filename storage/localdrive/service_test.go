package localdrive

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/ayumi-otosaka-314/brec-pp/storage"
)

func Test_service_GetRemovables(t *testing.T) {
	t.Parallel()

	s := &service{
		logger:   zaptest.NewLogger(t),
		rootPath: createTempFiles(t),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	removables, err := s.GetRemovables(ctx)
	require.NoError(t, err)

	count := 0
	for remove := range removables {
		size, err := remove()
		require.NoError(t, err)
		t.Log("removed and ensured size", size)
		count++
	}
	assert.Equal(t, 4, count)
}

func Test_service_GetRemovables_emptyDir(t *testing.T) {
	t.Parallel()

	s := &service{
		logger:   zaptest.NewLogger(t),
		rootPath: path.Join(createTempFiles(t), "emptyDir"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	removables, err := s.GetRemovables(ctx)
	require.NoError(t, err)

	count := 0
	for remove := range removables {
		size, err := remove()
		require.NoError(t, err)
		t.Log("removed and ensured size", size)
		count++
	}
	assert.Equal(t, 0, count) // no entries should be returned
}

func TestEnsureLocalCapacity(t *testing.T) {
	testPath := createTempFiles(t)
	s := &service{
		logger:   zaptest.NewLogger(t),
		rootPath: testPath,
	}

	currentCapacity, err := s.GetAvailableCapacity()
	require.NoError(t, err)
	assert.NoError(t, storage.EnsureCapacity(
		context.Background(),
		currentCapacity+4, // only first file with size 5 should be deleted
		s,
	))
}

func createTempFiles(t *testing.T) string {
	t.Helper()
	testPath := t.TempDir()

	file1, err := os.Create(path.Join(testPath, "test1"))
	require.NoError(t, err)
	file1.WriteString("test1")
	file1.Close()

	file2, err := os.Create(path.Join(testPath, "test2"))
	require.NoError(t, err)
	file2.WriteString("test2test2")
	file2.Close()

	require.NoError(t, os.Mkdir(path.Join(testPath, "emptyDir"), 0775))

	require.NoError(t, os.Mkdir(path.Join(testPath, "nonEmptyDir"), 0775))
	file3, err := os.Create(path.Join(testPath, "nonEmptyDir", "test3"))
	require.NoError(t, err)
	file3.WriteString("test3test3test3")
	file3.Close()

	return testPath
}

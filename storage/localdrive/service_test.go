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
	s := &service{
		logger:   zaptest.NewLogger(t),
		rootPath: createTempFiles(t),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	removables, err := s.GetRemovables(ctx)
	require.NoError(t, err)

	for remove := range removables {
		size, err := remove()
		require.NoError(t, err)
		t.Log("removed and ensured size", size)
	}
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

	return testPath
}

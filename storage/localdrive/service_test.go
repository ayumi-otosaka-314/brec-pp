package localdrive

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func Test_service_GetRemovables(t *testing.T) {
	testPath := t.TempDir()
	s := &service{
		logger:   zaptest.NewLogger(t),
		rootPath: testPath,
	}

	file1, err := os.Create(path.Join(testPath, "test1"))
	require.NoError(t, err)
	file1.WriteString("test1")
	file1.Close()

	file2, err := os.Create(path.Join(testPath, "test2"))
	require.NoError(t, err)
	file2.WriteString("test2test2")
	file2.Close()

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

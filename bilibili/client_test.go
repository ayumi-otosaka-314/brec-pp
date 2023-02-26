//go:build testRealAPI

package bilibili

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func Test_client_GetLiveInfo(t *testing.T) {
	c := NewClient(zaptest.NewLogger(t))
	got, err := c.GetLiveInfo(context.Background(), 1)
	assert.NoError(t, err)
	if !assert.NotNil(t, got.Data) {
		t.FailNow()
	}
	assert.NotNil(t, got.Data.Room)
	assert.NotNil(t, got.Data.Streamer)
}

package testutil

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func SetupLogger(t *testing.T) (*logrus.Logger, *test.Hook) {
	t.Helper()
	logger, hook := test.NewNullLogger()

	// reset logger after test
	t.Cleanup(func() {
		hook.Reset()
	})

	return logger, hook
}

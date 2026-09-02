package server

import (
	"bytes"
	"runtime/pprof"
	"testing"

	"github.com/stretchr/testify/require"
)

func requireNoGoroutineLeaks(t *testing.T) {
	t.Helper()

	profile := pprof.Lookup("goroutineleak")
	require.NotNil(t, profile)

	if n := profile.Count(); n > 0 {
		var buf bytes.Buffer
		_ = profile.WriteTo(&buf, 2)
		t.Logf("goroutine leak detected (%d):\n%s", n, buf.String())
		require.FailNow(t, "goroutine leak detected")
	}
}

func withLeakCheck(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { requireNoGoroutineLeaks(t) })
}

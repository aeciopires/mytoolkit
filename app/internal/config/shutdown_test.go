package config

import (
	"testing"
	"time"
)

func TestShutdownDelay(t *testing.T) {
	for in, want := range map[string]time.Duration{
		"": 0, "5s": 5 * time.Second, "250ms": 250 * time.Millisecond, "bogus": 0, "-3s": 0,
	} {
		t.Setenv(EnvShutdownDelay, in)
		if got := ShutdownDelay(); got != want {
			t.Errorf("ShutdownDelay(%q) = %v, want %v", in, got, want)
		}
	}
}

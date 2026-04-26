package watch

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestWatch_CallsFnMultipleTimes(t *testing.T) {
	count := 0
	done := make(chan error, 1)
	go func() {
		done <- Watch(50*time.Millisecond, func() error {
			count++
			if count >= 2 {
				syscall.Kill(os.Getpid(), syscall.SIGINT) //nolint:errcheck
			}
			return nil
		})
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Watch returned unexpected error: %v", err)
		}
		if count < 2 {
			t.Errorf("fn called %d times, want ≥ 2", count)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Watch did not return within 5s")
	}
}

func TestWatch_PropagatesFnError(t *testing.T) {
	sentinel := errors.New("fn failed")
	err := Watch(50*time.Millisecond, func() error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Watch returned %v, want %v", err, sentinel)
	}
}

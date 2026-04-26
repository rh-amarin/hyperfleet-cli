package watch

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Watch clears the terminal, calls fn, prints a footer, and repeats every interval.
// Returns nil when interrupted by SIGINT or SIGTERM; propagates fn errors immediately.
func Watch(interval time.Duration, fn func() error) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	for {
		fmt.Print("\033[H\033[2J")
		if err := fn(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "\nLast updated: %s  (Ctrl+C to stop)\n", time.Now().Format("15:04:05"))

		select {
		case <-quit:
			return nil
		case <-time.After(interval):
		}
	}
}

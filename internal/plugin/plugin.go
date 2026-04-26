package plugin

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Find searches PATH for an executable named "hf-<name>".
// Returns the resolved absolute path and true if found.
func Find(name string) (string, bool) {
	binary := "hf-" + strings.ToLower(name)
	path, err := exec.LookPath(binary)
	return path, err == nil
}

// Exec replaces the current process with the plugin binary, forwarding args.
// On POSIX systems this calls syscall.Exec and never returns on success.
// Falls back to exec.Command on platforms that don't support syscall.Exec.
func Exec(path string, args []string) error {
	argv := append([]string{path}, args...)
	return syscall.Exec(path, argv, os.Environ())
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const activeEnvFile = ".active-env"

// Store manages file-based configuration under a single directory.
type Store struct {
	dir       string
	activeEnv string
}

// EnvProfile describes a named environment profile.
type EnvProfile struct {
	Name       string
	PropCount  int
	IsActive   bool
}

// EnvEntry is a key/value pair annotated with whether an env overrides it.
type EnvEntry struct {
	Entry
	Value      string
	FromEnv    bool // true when the value comes from the named env profile
}

// NewStore returns a Store rooted at dir, creating it if needed.
// dir defaults to ~/.config/hf when empty.
func NewStore(dir string) (*Store, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		dir = filepath.Join(home, ".config", "hf")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("cannot create config dir %s: %w", dir, err)
	}
	s := &Store{dir: dir}
	s.activeEnv = s.readFile(activeEnvFile)
	return s, nil
}

// Dir returns the config directory path.
func (s *Store) Dir() string { return s.dir }

// ActiveEnv returns the name of the currently active env profile, or "".
func (s *Store) ActiveEnv() string { return s.activeEnv }

// Get returns the value for key following the precedence chain:
//   env var > active env profile file > base file > default
func (s *Store) Get(key string) string {
	e, ok := LookupEntry(key)
	if !ok {
		return s.readFile(key)
	}
	// 1. HF_* env var
	if e.EnvVar != "" {
		if v := os.Getenv(e.EnvVar); v != "" {
			return v
		}
	}
	// 2. active env profile file
	if s.activeEnv != "" {
		if v := s.readFile(s.activeEnv + "." + key); v != "" {
			return v
		}
	}
	// 3. base file
	if v := s.readFile(key); v != "" {
		return v
	}
	// 4. default — special-case registry key
	if key == "registry" && e.Default == "" {
		return os.Getenv("USER")
	}
	return e.Default
}

// Set writes value to ~/.config/hf/<key>.
func (s *Store) Set(key, value string) error {
	return os.WriteFile(s.filePath(key), []byte(value), 0o600)
}

// Clear deletes ~/.config/hf/<key>. Returns nil if the file does not exist.
func (s *Store) Clear(key string) error {
	err := os.Remove(s.filePath(key))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// ClearAll deletes all base key files (not env profile files, not dot-files).
func (s *Store) ClearAll() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, de := range entries {
		name := de.Name()
		if de.IsDir() || strings.HasPrefix(name, ".") || strings.Contains(name, ".") {
			continue // skip dirs, hidden files, and env profile files (<env>.<key>)
		}
		if err := os.Remove(filepath.Join(s.dir, name)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// EnvList returns all environment profiles found in the config directory.
func (s *Store) EnvList() ([]EnvProfile, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, de := range entries {
		if de.IsDir() || strings.HasPrefix(de.Name(), ".") {
			continue
		}
		parts := strings.SplitN(de.Name(), ".", 2)
		if len(parts) == 2 {
			counts[parts[0]]++
		}
	}
	var profiles []EnvProfile
	for name, count := range counts {
		profiles = append(profiles, EnvProfile{
			Name:      name,
			PropCount: count,
			IsActive:  name == s.activeEnv,
		})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return profiles, nil
}

// EnvActivate copies all <name>.<key> files over the base <key> files and
// writes the active env pointer.
func (s *Store) EnvActivate(name string) error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	prefix := name + "."
	found := false
	for _, de := range entries {
		if de.IsDir() || !strings.HasPrefix(de.Name(), prefix) {
			continue
		}
		found = true
		key := strings.TrimPrefix(de.Name(), prefix)
		data, err := os.ReadFile(filepath.Join(s.dir, de.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(s.filePath(key), data, 0o600); err != nil {
			return err
		}
	}
	if !found {
		return fmt.Errorf("environment %q not found in %s", name, s.dir)
	}
	if err := os.WriteFile(s.filePath(activeEnvFile), []byte(name), 0o600); err != nil {
		return err
	}
	s.activeEnv = name
	return nil
}

// EnvShow returns all registry entries annotated with whether the named env
// profile overrides the base value.
func (s *Store) EnvShow(name string) []EnvEntry {
	var out []EnvEntry
	for _, e := range Registry {
		envVal := s.readFile(name + "." + e.Key)
		baseVal := s.readFile(e.Key)
		value := baseVal
		fromEnv := false
		if envVal != "" {
			value = envVal
			fromEnv = true
		}
		if value == "" {
			value = e.Default
		}
		out = append(out, EnvEntry{Entry: e, Value: value, FromEnv: fromEnv})
	}
	return out
}

// — helpers —

func (s *Store) filePath(name string) string {
	return filepath.Join(s.dir, name)
}

func (s *Store) readFile(name string) string {
	data, err := os.ReadFile(s.filePath(name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

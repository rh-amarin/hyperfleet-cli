package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configFile = "config.yaml"
	stateFile  = "state.yaml"
	envsDir    = "environments"
)

// Store holds the fully resolved configuration and provides persistence.
type Store struct {
	dir   string
	cfg   Config // merged: defaults ← config.yaml ← active env profile
	state State
	// raw tracks what came from config.yaml (before env merge) for source annotation
	rawCfg Config
}

// NewStore loads or initialises configuration from dir (~/.config/hf by default).
func NewStore(dir string) (*Store, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		dir = filepath.Join(home, ".config", "hf")
	}

	for _, d := range []string{dir, filepath.Join(dir, envsDir)} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return nil, fmt.Errorf("cannot create directory %s: %w", d, err)
		}
	}

	s := &Store{dir: dir}

	// Load config.yaml once; unmarshal into both cfg (with defaults) and rawCfg (zero base).
	s.cfg = defaults()
	cfgPath := filepath.Join(dir, configFile)
	cfgData, err := s.readOrCreate(cfgPath, &s.rawCfg)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(cfgData, &s.cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", cfgPath, err)
	}

	// Load state.yaml (create empty if missing).
	if err := s.loadYAML(stateFile, &s.state); err != nil {
		return nil, err
	}

	// Deep-merge active env profile on top of cfg.
	if s.state.ActiveEnvironment != "" {
		envFile := filepath.Join(dir, envsDir, s.state.ActiveEnvironment+".yaml")
		if err := s.loadYAML(envFile, &s.cfg); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading env profile %q: %w", s.state.ActiveEnvironment, err)
		}
	}

	return s, nil
}

// Dir returns the config directory path.
func (s *Store) Dir() string { return s.dir }

// Cfg returns the fully resolved Config (defaults ← config.yaml ← env profile).
func (s *Store) Cfg() Config { return s.cfg }

// RawCfg returns the config as read from config.yaml (before env merge).
func (s *Store) RawCfg() Config { return s.rawCfg }

// SetRawCfg replaces rawCfg and saves config.yaml.
func (s *Store) SetRawCfg(cfg Config) error {
	s.rawCfg = cfg
	return s.Save()
}

// State returns the current State.
func (s *Store) State() State { return s.state }

// Save writes the current rawCfg (un-env-merged) back to config.yaml.
func (s *Store) Save() error {
	return s.writeYAML(configFile, &s.rawCfg)
}

// SetState updates one state field by key name and atomically writes state.yaml.
// Valid keys: active-environment, cluster-id, cluster-name, nodepool-id.
func (s *Store) SetState(key, val string) error {
	switch key {
	case "active-environment":
		s.state.ActiveEnvironment = val
	case "cluster-id":
		s.state.ClusterID = val
	case "cluster-name":
		s.state.ClusterName = val
	case "nodepool-id":
		s.state.NodePoolID = val
	default:
		return fmt.Errorf("unknown state key %q", key)
	}
	return s.writeYAML(stateFile, &s.state)
}

// ClearState zeros all runtime state fields and writes state.yaml.
func (s *Store) ClearState() error {
	s.state = State{}
	return s.writeYAML(stateFile, &s.state)
}

// SetConfigValue updates a field in rawCfg by dotted path and saves config.yaml.
// Path format: "<section>.<key>", e.g. "hyperfleet.api-url".
func (s *Store) SetConfigValue(path, value string) error {
	if err := setField(&s.rawCfg, path, value); err != nil {
		return err
	}
	return s.Save()
}

// ClearConfigValue resets a field in rawCfg to its default value and saves.
func (s *Store) ClearConfigValue(path string) error {
	def := defaults()
	val, err := getField(&def, path)
	if err != nil {
		return err
	}
	return s.SetConfigValue(path, val)
}

// OverrideCfg applies a flag-level override to the in-memory cfg only (not saved to disk).
func (s *Store) OverrideCfg(path, value string) error {
	return setField(&s.cfg, path, value)
}

// EnvCfg returns a fully merged Config as if the named env were active.
func (s *Store) EnvCfg(name string) (Config, error) {
	cfg := defaults()
	if err := s.loadYAMLReadOnly(filepath.Join(s.dir, envsDir, name+".yaml"), &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// SaveEnv writes envCfg to environments/<name>.yaml.
func (s *Store) SaveEnv(name string, envCfg *Config) error {
	path := filepath.Join(s.dir, envsDir, name+".yaml")
	return s.writeYAMLPath(path, envCfg)
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (s *Store) loadYAMLReadOnly(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

// readOrCreate reads path and unmarshals into out; if the file is missing it
// writes out's current value as the initial file and returns empty bytes.
func (s *Store) readOrCreate(path string, out interface{}) ([]byte, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, s.writeYAMLPath(path, out)
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return data, nil
}

func (s *Store) loadYAML(name string, out interface{}) error {
	path := name
	if !filepath.IsAbs(name) {
		path = filepath.Join(s.dir, name)
	}
	_, err := s.readOrCreate(path, out)
	return err
}

func (s *Store) writeYAML(name string, v interface{}) error {
	return s.writeYAMLPath(filepath.Join(s.dir, name), v)
}

func (s *Store) writeYAMLPath(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}



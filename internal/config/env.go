package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// EnvProfile describes a named environment profile.
type EnvProfile struct {
	Name      string
	PropCount int
	IsActive  bool
}

// EnvList scans the environments/ directory and returns all profiles.
func (s *Store) EnvList() ([]EnvProfile, error) {
	dir := filepath.Join(s.dir, envsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var profiles []EnvProfile
	for _, de := range entries {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(de.Name(), ".yaml")
		count := s.envPropCount(filepath.Join(dir, de.Name()))
		profiles = append(profiles, EnvProfile{
			Name:      name,
			PropCount: count,
			IsActive:  name == s.state.ActiveEnvironment,
		})
	}
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })
	return profiles, nil
}

// EnvActivate sets active-environment in state.yaml to name.
func (s *Store) EnvActivate(name string) error {
	path := filepath.Join(s.dir, envsDir, name+".yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.ErrNotExist
	}
	return s.SetState("active-environment", name)
}

// EnvDeactivate clears active-environment from state.yaml.
func (s *Store) EnvDeactivate() error {
	return s.SetState("active-environment", "")
}

// EnvShow returns resolved values annotated for a specific env profile.
// It merges: defaults ← config.yaml ← env profile, then returns with per-key source.
func (s *Store) EnvShow(name string) ([]ResolvedValue, error) {
	// Temporarily activate the named env for resolution.
	saved := s.state.ActiveEnvironment
	s.state.ActiveEnvironment = name
	resolved := s.Resolve()
	s.state.ActiveEnvironment = saved
	return resolved, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

// envPropCount counts non-zero/non-default keys in the YAML file.
func (s *Store) envPropCount(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return 0
	}
	count := 0
	for _, v := range raw {
		if m, ok := v.(map[string]interface{}); ok {
			count += len(m)
		}
	}
	return count
}

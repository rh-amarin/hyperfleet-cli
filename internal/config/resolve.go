package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ResolvedValue holds a config value and its source annotation.
type ResolvedValue struct {
	Path    string
	Section string
	Value   string
	Source  string // [default], [config], [env:<name>]
}

// Resolve returns the value and source for every config path.
func (s *Store) Resolve() []ResolvedValue {
	def := defaults()
	rawFileValues := s.rawCfg
	activeEnv := s.state.ActiveEnvironment

	// Load env profile for comparison.
	envCfg := Config{}
	if activeEnv != "" {
		envFile := filepath.Join(s.dir, envsDir, activeEnv+".yaml")
		if data, err := os.ReadFile(envFile); err == nil {
			_ = yaml.Unmarshal(data, &envCfg)
		}
	}

	var out []ResolvedValue
	for _, p := range AllPaths {
		defVal, _ := getField(&def, p.Path)
		fileVal, _ := getField(&rawFileValues, p.Path)
		envProfileVal, _ := getField(&envCfg, p.Path)

		value := defVal
		source := "[default]"

		if fileVal != "" && fileVal != defVal {
			value = fileVal
			source = "[config]"
		}
		if activeEnv != "" && envProfileVal != "" && envProfileVal != defVal {
			value = envProfileVal
			source = "[env:" + activeEnv + "]"
		}

		out = append(out, ResolvedValue{
			Path:    p.Path,
			Section: p.Section,
			Value:   value,
			Source:  source,
		})
	}
	return out
}

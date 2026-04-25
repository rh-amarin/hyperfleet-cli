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
	Source  string // [default], [config], [env:<name>], [ENV], [flag]
}

// envVarFor maps dotted paths to their HF_* override variable.
var envVarFor = map[string]string{
	"hyperfleet.api-url":     "HF_API_URL",
	"hyperfleet.api-version": "HF_API_VERSION",
	"hyperfleet.token":       "HF_TOKEN",
	"kubernetes.context":     "HF_CONTEXT",
	"kubernetes.namespace":   "HF_NAMESPACE",
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
		envVar := envVarFor[p.Path]
		envVarVal := ""
		if envVar != "" {
			envVarVal = os.Getenv(envVar)
		}

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
		if envVarVal != "" {
			value = envVarVal
			source = "[ENV]"
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

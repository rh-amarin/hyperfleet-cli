package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/spf13/cobra"
)

// ── ANSI colours ─────────────────────────────────────────────────────────────

const (
	ansiReset  = "\033[0m"
	ansiGreen  = "\033[0;32m"
	ansiCyan   = "\033[0;36m"
	ansiBold   = "\033[1m"
)

func colour(code, s string) string {
	if noColor {
		return s
	}
	return code + s + ansiReset
}

// ── Root config command ───────────────────────────────────────────────────────

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage HyperFleet CLI configuration",
	Long: `Manage HyperFleet CLI configuration.

With no subcommand, prints usage, environment list, and current configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printConfigUsage()
		fmt.Println()
		if err := printEnvList(store); err != nil {
			return err
		}
		fmt.Println()
		printConfigSections(store, "")
		return nil
	},
}

// ── show ──────────────────────────────────────────────────────────────────────

var configShowCmd = &cobra.Command{
	Use:   "show [env-name]",
	Short: "Show current configuration",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		env := ""
		if len(args) > 0 {
			env = args[0]
		}
		if err := printEnvList(store); err != nil {
			return err
		}
		fmt.Println()
		printConfigSections(store, env)
		return nil
	},
}

// ── set ───────────────────────────────────────────────────────────────────────

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if _, ok := config.LookupEntry(key); !ok {
			fmt.Fprintf(os.Stderr, "[WARN] unknown config key %q — setting anyway\n", key)
		}
		if err := store.Set(key, value); err != nil {
			return err
		}
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

// ── clear ─────────────────────────────────────────────────────────────────────

var configClearCmd = &cobra.Command{
	Use:   "clear <key|all>",
	Short: "Clear a configuration value (or all values)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] == "all" {
			if err := store.ClearAll(); err != nil {
				return err
			}
			fmt.Println("Cleared all configuration values.")
			return nil
		}
		if err := store.Clear(args[0]); err != nil {
			return err
		}
		fmt.Printf("Cleared %s\n", args[0])
		return nil
	},
}

// ── doctor ────────────────────────────────────────────────────────────────────

var configDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check which commands have all required config set",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(colour(ansiBold, "Config Doctor"))
		// Required keys per command group — mirrors hf_require_config usage in scripts.
		type commandCheck struct {
			name string
			keys []string
		}
		checks := []commandCheck{
			{"hf cluster create", []string{"api-url", "api-version"}},
			{"hf cluster get/list/search/delete", []string{"api-url", "api-version", "cluster-id"}},
			{"hf cluster patch", []string{"api-url", "api-version", "cluster-id"}},
			{"hf cluster conditions/statuses", []string{"api-url", "api-version", "cluster-id"}},
			{"hf cluster adapter post-status", []string{"api-url", "api-version", "cluster-id"}},
			{"hf nodepool create", []string{"api-url", "api-version", "cluster-id"}},
			{"hf nodepool get/list/search/delete", []string{"api-url", "api-version", "cluster-id", "nodepool-id"}},
			{"hf nodepool adapter post-status", []string{"api-url", "api-version", "cluster-id", "nodepool-id"}},
			{"hf db query/delete/statuses", []string{"db-host", "db-port", "db-name", "db-user", "db-password"}},
			{"hf maestro list/get/delete", []string{"maestro-http-endpoint", "maestro-consumer"}},
			{"hf kube port-forward", []string{"context", "namespace", "pf-api-port", "pf-pg-port"}},
			{"hf logs", []string{"context", "namespace"}},
			{"hf pubsub publish", []string{"gcp-project", "cluster-id"}},
			{"hf rabbitmq publish", []string{"rabbitmq-host", "rabbitmq-mgmt-port", "rabbitmq-user", "rabbitmq-password"}},
		}
		for _, c := range checks {
			missing := []string{}
			for _, k := range c.keys {
				if store.Get(k) == "" {
					missing = append(missing, k)
				}
			}
			if len(missing) == 0 {
				fmt.Printf("  %s %s\n", colour(ansiGreen, "●"), c.name)
			} else {
				fmt.Printf("  ○ %s — missing: %s\n", c.name, strings.Join(missing, ", "))
			}
		}
		return nil
	},
}

// ── bootstrap ─────────────────────────────────────────────────────────────────

var configBootstrapCmd = &cobra.Command{
	Use:   "bootstrap [env-name]",
	Short: "Interactive environment setup",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		envName := ""
		if len(args) > 0 {
			envName = args[0]
		}
		scanner := bufio.NewScanner(os.Stdin)
		for _, section := range config.Sections() {
			fmt.Printf("\n%s\n", colour(ansiBold, section))
			for _, e := range config.EntriesForSection(section) {
				current := store.Get(e.Key)
				prompt := fmt.Sprintf("  %s [%s]: ", e.Key, current)
				fmt.Print(prompt)
				if !scanner.Scan() {
					break
				}
				input := strings.TrimSpace(scanner.Text())
				if input == "" {
					continue
				}
				key := e.Key
				if envName != "" {
					key = envName + "." + e.Key
				}
				if err := store.Set(key, input); err != nil {
					return err
				}
			}
		}
		if envName != "" {
			fmt.Printf("\nEnvironment %q saved. Activate with: hf config env activate %s\n", envName, envName)
		} else {
			fmt.Println("\nConfiguration saved.")
		}
		return nil
	},
}

// ── env ───────────────────────────────────────────────────────────────────────

var configEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment profiles",
}

var configEnvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List environment profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		return printEnvList(store)
	},
}

var configEnvShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show config with environment overrides highlighted",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		printConfigSections(store, args[0])
		return nil
	},
}

var configEnvActivateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Activate an environment profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := store.EnvActivate(name); err != nil {
			return err
		}
		fmt.Printf("Activated environment %q\n", name)
		return nil
	},
}

// ── helpers ───────────────────────────────────────────────────────────────────

func printConfigUsage() {
	fmt.Println(`Usage: hf config [show|set|clear|doctor|bootstrap|env] [args...]

Commands:
  show [env-name]         Show current configuration
  set <key> <value>       Set a configuration value
  clear <key>             Clear a configuration value
  clear all               Clear all configuration
  doctor                  Check which commands are ready to use
  bootstrap [env-name]    Interactive environment setup
  env list                List environments
  env show <name>         Show config with environment overrides
  env activate <name>     Activate an environment`)
}

func printEnvList(s *config.Store) error {
	profiles, err := s.EnvList()
	if err != nil {
		return err
	}
	fmt.Println(colour(ansiBold, "Environments"))
	if len(profiles) == 0 {
		fmt.Println("  (none)")
		return nil
	}
	for _, p := range profiles {
		marker := "○"
		line := fmt.Sprintf("  %s %-32s %d properties", marker, p.Name, p.PropCount)
		if p.IsActive {
			line = fmt.Sprintf("  %s %-32s %d properties (active)",
				colour(ansiGreen, "●"), p.Name, p.PropCount)
		}
		fmt.Println(line)
	}
	return nil
}

func printConfigSections(s *config.Store, envName string) {
	fmt.Println(colour(ansiBold, "HyperFleet Configuration"))
	fmt.Printf("  Config dir: %s\n", s.Dir())
	if active := s.ActiveEnv(); active != "" {
		fmt.Printf("  Active env: %s\n", colour(ansiGreen, active))
	}
	fmt.Println()

	var entries []config.EnvEntry
	if envName != "" {
		entries = s.EnvShow(envName)
	} else {
		// Build entries showing current effective values.
		for _, e := range config.Registry {
			val := s.Get(e.Key)
			fromEnv := false
			if s.ActiveEnv() != "" {
				// Check if the active env file provides the value.
				// We detect this by seeing if Get differs from the base file.
				_ = fromEnv // annotation below
			}
			entries = append(entries, config.EnvEntry{Entry: e, Value: val})
		}
	}

	currentSection := ""
	for _, ee := range entries {
		if ee.Section != currentSection {
			currentSection = ee.Section
			fmt.Printf("%s\n", colour(ansiBold, currentSection))
		}
		displayVal := ee.Value
		if ee.IsSecret {
			if ee.Value != "" {
				displayVal = "<set>"
			} else {
				displayVal = "<not set>"
			}
		}
		line := fmt.Sprintf("  %-36s %s", ee.Key, displayVal)
		if ee.FromEnv && envName != "" {
			line += " " + colour(ansiCyan, "["+envName+"]")
		}
		fmt.Println(line)
	}
}

// ── registration ──────────────────────────────────────────────────────────────

func init() {
	configEnvCmd.AddCommand(configEnvListCmd, configEnvShowCmd, configEnvActivateCmd)
	configCmd.AddCommand(
		configShowCmd,
		configSetCmd,
		configClearCmd,
		configDoctorCmd,
		configBootstrapCmd,
		configEnvCmd,
	)
	rootCmd.AddCommand(configCmd)
}

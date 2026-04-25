package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/spf13/cobra"
)

// ── ANSI helpers ──────────────────────────────────────────────────────────────

func cyan(s string) string {
	if noColor {
		return s
	}
	return "\033[36m" + s + "\033[0m"
}

func bold(s string) string {
	if noColor {
		return s
	}
	return "\033[1m" + s + "\033[0m"
}

func dim(s string) string {
	if noColor {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

// ── hf config ─────────────────────────────────────────────────────────────────

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "View and modify hf configuration, environment profiles, and runtime state.",
	RunE: func(cmd *cobra.Command, args []string) error {
		printEnvList()
		fmt.Println()
		return printConfigShow("")
	},
}

// ── hf config show [env] ──────────────────────────────────────────────────────

var configShowCmd = &cobra.Command{
	Use:   "show [env]",
	Short: "Show resolved configuration with source annotations",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		env := ""
		if len(args) == 1 {
			env = args[0]
		}
		return printConfigShow(env)
	},
}

func printConfigShow(env string) error {
	var resolved []config.ResolvedValue
	if env != "" {
		var err error
		resolved, err = cfgStore.EnvShow(env)
		if err != nil {
			return fmt.Errorf("cannot show env %q: %w", env, err)
		}
	} else {
		resolved = cfgStore.Resolve()
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	currentSection := ""
	for _, rv := range resolved {
		if rv.Section != currentSection {
			if currentSection != "" {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s\n", bold("["+rv.Section+"]"))
			currentSection = rv.Section
		}
		key := strings.TrimPrefix(rv.Path, rv.Section+".")
		displayVal := rv.Value
		if config.IsSecret(rv.Path) {
			if rv.Value != "" {
				displayVal = "<set>"
			} else {
				displayVal = "<not set>"
			}
		}
		src := dim(rv.Source)
		// Highlight env-sourced values in cyan
		if strings.HasPrefix(rv.Source, "[env:") {
			displayVal = cyan(displayVal)
			src = cyan(rv.Source)
		}
		fmt.Fprintf(w, "  %-38s\t%s\t%s\n", key, displayVal, src)
	}
	w.Flush()

	// Print state summary
	state := cfgStore.State()
	fmt.Println()
	fmt.Println(bold("state:"))
	fmt.Printf("  active-environment  %s\n", orDash(state.ActiveEnvironment))
	fmt.Printf("  cluster-id          %s\n", orDash(state.ClusterID))
	fmt.Printf("  cluster-name        %s\n", orDash(state.ClusterName))
	fmt.Printf("  nodepool-id         %s\n", orDash(state.NodePoolID))
	return nil
}

func orDash(s string) string {
	if s == "" {
		return dim("-")
	}
	return s
}

// ── hf config set ─────────────────────────────────────────────────────────────

var configSetCmd = &cobra.Command{
	Use:   "set <section.key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgStore.SetConfigValue(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("set %s = %s\n", args[0], args[1])
		return nil
	},
}

// ── hf config clear ───────────────────────────────────────────────────────────

var configClearCmd = &cobra.Command{
	Use:   "clear <section.key|state>",
	Short: "Reset a config field to default, or clear all runtime state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] == "state" {
			if err := cfgStore.ClearState(); err != nil {
				return err
			}
			fmt.Println("runtime state cleared")
			return nil
		}
		if err := cfgStore.ClearConfigValue(args[0]); err != nil {
			return err
		}
		fmt.Printf("reset %s to default\n", args[0])
		return nil
	},
}

// ── hf config doctor ──────────────────────────────────────────────────────────

var configDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check configuration readiness per command group",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cfgStore.Cfg()

		checks := []struct {
			group  string
			checks []struct {
				label string
				ok    bool
			}
		}{
			{"cluster / nodepool / adapters", []struct {
				label string
				ok    bool
			}{
				{"hyperfleet.api-url", cfg.Hyperfleet.APIURL != ""},
				{"hyperfleet.token", cfg.Hyperfleet.Token != ""},
			}},
			{"kubernetes", []struct {
				label string
				ok    bool
			}{
				{"kubernetes.context", cfg.Kubernetes.Context != ""},
				{"kubernetes.namespace", cfg.Kubernetes.Namespace != ""},
			}},
			{"database", []struct {
				label string
				ok    bool
			}{
				{"database.host", cfg.Database.Host != ""},
				{"database.user", cfg.Database.User != ""},
				{"database.name", cfg.Database.Name != ""},
			}},
			{"rabbitmq", []struct {
				label string
				ok    bool
			}{
				{"rabbitmq.host", cfg.RabbitMQ.Host != ""},
				{"rabbitmq.user", cfg.RabbitMQ.User != ""},
			}},
			{"registry", []struct {
				label string
				ok    bool
			}{
				{"registry.name", cfg.Registry.Name != ""},
			}},
		}

		for _, g := range checks {
			allOK := true
			for _, c := range g.checks {
				if !c.ok {
					allOK = false
					break
				}
			}
			indicator := "●"
			if !allOK {
				indicator = "○"
			}
			fmt.Printf("%s  %s\n", indicator, bold(g.group))
			for _, c := range g.checks {
				mark := "✓"
				if !c.ok {
					mark = "✗"
				}
				fmt.Printf("     %s  %s\n", mark, c.label)
			}
		}
		return nil
	},
}

// ── hf config bootstrap ───────────────────────────────────────────────────────

var configBootstrapCmd = &cobra.Command{
	Use:   "bootstrap [env-name]",
	Short: "Interactively create a config or environment profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		envName := ""
		if len(args) == 1 {
			envName = args[0]
		}

		scanner := bufio.NewScanner(os.Stdin)
		cfg := cfgStore.Cfg()
		var target config.Config
		isEnvMode := envName != ""

		// In global config mode, blank = keep current value.
		// In env profile mode, blank = omit (inherit from config.yaml/defaults).
		prompt := func(label, hint string) string {
			if hint != "" {
				fmt.Printf("  %s [%s]: ", label, hint)
			} else {
				fmt.Printf("  %s: ", label)
			}
			if !scanner.Scan() {
				if isEnvMode {
					return ""
				}
				return hint
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				if isEnvMode {
					return ""
				}
				return hint
			}
			return line
		}

		if isEnvMode {
			fmt.Printf("Bootstrapping environment profile %q\n", envName)
			fmt.Println("(Leave blank to inherit from config.yaml or defaults)")
		} else {
			fmt.Println("Bootstrapping hf configuration")
			fmt.Println("(Leave blank to keep current value)")
			target = cfgStore.RawCfg()
		}
		fmt.Println()

		// Hyperfleet section
		fmt.Println(bold("[hyperfleet]"))
		target.Hyperfleet.APIURL = prompt("api-url", cfg.Hyperfleet.APIURL)
		target.Hyperfleet.Token = prompt("token", "")
		target.Hyperfleet.GCPProject = prompt("gcp-project", cfg.Hyperfleet.GCPProject)
		fmt.Println()

		// Kubernetes section
		fmt.Println(bold("[kubernetes]"))
		target.Kubernetes.Context = prompt("context", cfg.Kubernetes.Context)
		target.Kubernetes.Namespace = prompt("namespace", cfg.Kubernetes.Namespace)
		fmt.Println()

		// Database section
		fmt.Println(bold("[database]"))
		target.Database.Host = prompt("host", cfg.Database.Host)
		target.Database.User = prompt("user", cfg.Database.User)
		target.Database.Name = prompt("name", cfg.Database.Name)
		target.Database.Password = prompt("password", "")
		fmt.Println()

		// RabbitMQ section
		fmt.Println(bold("[rabbitmq]"))
		target.RabbitMQ.Host = prompt("host", cfg.RabbitMQ.Host)
		target.RabbitMQ.User = prompt("user", cfg.RabbitMQ.User)
		target.RabbitMQ.Password = prompt("password", "")
		fmt.Println()

		if envName != "" {
			if err := cfgStore.SaveEnv(envName, &target); err != nil {
				return fmt.Errorf("saving env profile: %w", err)
			}
			fmt.Printf("Saved environment profile to environments/%s.yaml\n", envName)
		} else {
			// Merge filled values back into rawCfg
			if err := cfgStore.SetRawCfg(target); err != nil {
				return err
			}
			fmt.Println("Configuration saved to config.yaml")
		}
		return nil
	},
}

// ── hf config env ─────────────────────────────────────────────────────────────

var configEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment profiles",
}

var configEnvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		printEnvList()
		return nil
	},
}

func printEnvList() {
	profiles, err := cfgStore.EnvList()
	if err != nil || len(profiles) == 0 {
		fmt.Println(dim("no environment profiles found"))
		return
	}
	fmt.Println(bold("environments:"))
	for _, p := range profiles {
		active := "  "
		name := p.Name
		if p.IsActive {
			active = "* "
			name = cyan(name)
		}
		fmt.Printf("  %s%-20s  %d key(s)\n", active, name, p.PropCount)
	}
}

var configEnvShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show resolved config as if env profile were active",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return printConfigShow(args[0])
	},
}

var configEnvActivateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Activate an environment profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgStore.EnvActivate(args[0]); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("environment profile %q not found", args[0])
			}
			return err
		}
		fmt.Printf("activated environment: %s\n", args[0])
		return nil
	},
}

var configEnvDeactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate the current environment profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfgStore.EnvDeactivate(); err != nil {
			return err
		}
		fmt.Println("environment deactivated")
		return nil
	},
}

// ── wiring ────────────────────────────────────────────────────────────────────

func init() {
	configEnvCmd.AddCommand(configEnvListCmd, configEnvShowCmd, configEnvActivateCmd, configEnvDeactivateCmd)
	configCmd.AddCommand(configShowCmd, configSetCmd, configClearCmd, configDoctorCmd, configBootstrapCmd, configEnvCmd)
	rootCmd.AddCommand(configCmd)
}

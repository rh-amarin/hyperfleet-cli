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

// ── active environment guard ──────────────────────────────────────────────────

func requireActiveEnv() error {
	if cfgStore.State().ActiveEnvironment == "" {
		return fmt.Errorf(
			"no active environment\n" +
				"  → run 'hf config env new' to create one\n" +
				"  → run 'hf config env activate <name>' to activate an existing one",
		)
	}
	return nil
}

// ── shared config table printer ───────────────────────────────────────────────

func printConfigTable(cfg *config.Config) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	currentSection := ""
	for _, p := range config.AllPaths {
		if p.Section != currentSection {
			if currentSection != "" {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s\n", bold("["+p.Section+"]"))
			currentSection = p.Section
		}
		key := strings.TrimPrefix(p.Path, p.Section+".")
		val, _ := config.GetField(cfg, p.Path)
		displayVal := secretDisplay(p.Path, val)
		fmt.Fprintf(w, "  %-38s\t%s\n", key, displayVal)
	}
	w.Flush()
}

func secretDisplay(path, val string) string {
	if !config.IsSecret(path) {
		return val
	}
	// Integer fields that are unset marshal as "0"; treat as not set.
	if val == "" || val == "0" {
		return "<not set>"
	}
	return "<set>"
}

// ── hf config ─────────────────────────────────────────────────────────────────

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "View and modify hf configuration, environment profiles, and runtime state.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireActiveEnv(); err != nil {
			return err
		}
		printEnvList()
		fmt.Println()
		return printConfigShow()
	},
}

// ── hf config show ────────────────────────────────────────────────────────────

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireActiveEnv(); err != nil {
			return err
		}
		return printConfigShow()
	},
}

func printConfigShow() error {
	fmt.Printf("active environment: %s\n\n", cyan(cfgStore.State().ActiveEnvironment))
	cfg := cfgStore.Cfg()
	printConfigTable(&cfg)

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
		if err := requireActiveEnv(); err != nil {
			return err
		}
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
		if err := requireActiveEnv(); err != nil {
			return err
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
		if err := requireActiveEnv(); err != nil {
			return err
		}
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
		cfg, err := cfgStore.EnvCfg(args[0])
		if err != nil {
			return fmt.Errorf("cannot show env %q: %w", args[0], err)
		}
		fmt.Printf("environment: %s\n\n", cyan(args[0]))
		printConfigTable(&cfg)
		return nil
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

// ── hf config env new ─────────────────────────────────────────────────────────

var configEnvNewCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new environment profile interactively",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := bufio.NewScanner(os.Stdin)

		promptLine := func(label, defaultVal string) string {
			if defaultVal != "" {
				fmt.Printf("  %s [%s]: ", label, defaultVal)
			} else {
				fmt.Printf("  %s: ", label)
			}
			if !scanner.Scan() {
				return ""
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				return ""
			}
			return line
		}

		envName := ""
		if len(args) == 1 {
			envName = args[0]
		} else {
			fmt.Printf("  environment name: ")
			if scanner.Scan() {
				envName = strings.TrimSpace(scanner.Text())
			}
			if envName == "" {
				return fmt.Errorf("environment name is required")
			}
		}

		fmt.Printf("\nConfiguring environment %q\n", envName)
		fmt.Println("(Leave blank to use the shown default)")
		fmt.Println()

		def := config.Defaults()
		var target config.Config

		fmt.Println(bold("[hyperfleet]"))
		target.Hyperfleet.APIURL = coalesce(promptLine("api-url", def.Hyperfleet.APIURL), def.Hyperfleet.APIURL)
		target.Hyperfleet.Token = promptLine("token", "")
		target.Hyperfleet.GCPProject = coalesce(promptLine("gcp-project", def.Hyperfleet.GCPProject), def.Hyperfleet.GCPProject)
		fmt.Println()

		fmt.Println(bold("[kubernetes]"))
		target.Kubernetes.Context = promptLine("context", "")
		target.Kubernetes.Namespace = promptLine("namespace", "")
		fmt.Println()

		fmt.Println(bold("[database]"))
		target.Database.Host = coalesce(promptLine("host", def.Database.Host), def.Database.Host)
		target.Database.User = coalesce(promptLine("user", def.Database.User), def.Database.User)
		target.Database.Name = coalesce(promptLine("name", def.Database.Name), def.Database.Name)
		target.Database.Password = coalesce(promptLine("password", def.Database.Password), def.Database.Password)
		fmt.Println()

		fmt.Println(bold("[rabbitmq]"))
		target.RabbitMQ.Host = coalesce(promptLine("host", def.RabbitMQ.Host), def.RabbitMQ.Host)
		target.RabbitMQ.User = coalesce(promptLine("user", def.RabbitMQ.User), def.RabbitMQ.User)
		target.RabbitMQ.Password = coalesce(promptLine("password", def.RabbitMQ.Password), def.RabbitMQ.Password)
		fmt.Println()

		if err := cfgStore.SaveEnv(envName, &target); err != nil {
			return fmt.Errorf("saving environment profile: %w", err)
		}
		fmt.Printf("Saved environment profile %q\n", envName)
		fmt.Printf("Run 'hf config env activate %s' to use it.\n", envName)
		return nil
	},
}

// coalesce returns a if non-empty, otherwise b.
func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ── wiring ────────────────────────────────────────────────────────────────────

func init() {
	configEnvCmd.AddCommand(configEnvListCmd, configEnvShowCmd, configEnvActivateCmd, configEnvDeactivateCmd, configEnvNewCmd)
	configCmd.AddCommand(configShowCmd, configSetCmd, configClearCmd, configDoctorCmd, configEnvCmd)
	rootCmd.AddCommand(configCmd)
}

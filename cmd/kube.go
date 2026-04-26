package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
	out "github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	kubeconfig    string
	kubeNamespace string
)

func init() {
	rootCmd.AddCommand(kubeCmd)

	kubeCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig (default: KUBECONFIG env, then ~/.kube/config)")
	kubeCmd.PersistentFlags().StringVarP(&kubeNamespace, "namespace", "n", "amarin-ns1", "Kubernetes namespace")

	// port-forward subcommands
	kubeCmd.AddCommand(kubePFCmd)
	kubePFCmd.AddCommand(kubePFStartCmd)
	kubePFCmd.AddCommand(kubePFStopCmd)
	kubePFCmd.AddCommand(kubePFStatusCmd)

	// other subcommands
	kubeCmd.AddCommand(kubeContextCmd)
	kubeCmd.AddCommand(kubeCurlCmd)
	kubeCmd.AddCommand(kubeDebugCmd)

	// hidden daemon command used by StartPortForward
	kubePFDaemonCmd.Hidden = true
	kubeCmd.AddCommand(kubePFDaemonCmd)
}

var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "Kubernetes operations: port-forward, logs, exec",
}

// ─── predefined port-forward services ────────────────────────────────────────

type predefinedPF struct {
	name       string // display name and PID file key
	podPattern string // pattern to find the target pod
	namespace  string
	localPort  int
	remotePort int
}

func predefinedPFs() []predefinedPF {
	apiPort := 8000
	pgPort := 5432
	maestroHTTPPort := 8100
	maestroHTTPRemote := 8000
	maestroGRPCPort := 8090
	maestroNs := "maestro"
	ns := kubeNamespace

	if cfgStore != nil {
		cfg := cfgStore.Cfg()
		if cfg.PortForward.APIPort != 0 {
			apiPort = cfg.PortForward.APIPort
		}
		if cfg.PortForward.PGPort != 0 {
			pgPort = cfg.PortForward.PGPort
		}
		if cfg.PortForward.MaestroHTTPPort != 0 {
			maestroHTTPPort = cfg.PortForward.MaestroHTTPPort
		}
		if cfg.PortForward.MaestroHTTPRemote != 0 {
			maestroHTTPRemote = cfg.PortForward.MaestroHTTPRemote
		}
		if cfg.PortForward.MaestroGRPCPort != 0 {
			maestroGRPCPort = cfg.PortForward.MaestroGRPCPort
		}
		if cfg.Maestro.Namespace != "" {
			maestroNs = cfg.Maestro.Namespace
		}
	}

	return []predefinedPF{
		{"hyperfleet-api", "hyperfleet-api", ns, apiPort, 8000},
		{"postgresql", "postgresql", ns, pgPort, 5432},
		{"maestro-http", "maestro", maestroNs, maestroHTTPPort, maestroHTTPRemote},
		{"maestro-grpc", "maestro", maestroNs, maestroGRPCPort, maestroGRPCPort},
	}
}

func predefinedByName(name string) (predefinedPF, bool) {
	for _, p := range predefinedPFs() {
		if p.name == name {
			return p, true
		}
	}
	return predefinedPF{}, false
}

// ─── context ──────────────────────────────────────────────────────────────────

var kubeContextCmd = &cobra.Command{
	Use:          "context",
	Short:        "Print the current kubeconfig context name",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := kube.CurrentContext(kubeconfig)
		if err != nil {
			return err
		}
		fmt.Println(ctx)
		return nil
	},
}

// ─── port-forward ─────────────────────────────────────────────────────────────

var kubePFCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Manage background port-forwards to in-cluster services",
}

// kubePFStartCmd: no args → start all predefined; one arg → start named predefined;
// two args → generic <service> <localPort:remotePort>.
var kubePFStartCmd = &cobra.Command{
	Use:          "start [name | <service> <localPort:remotePort>]",
	Short:        "Start port-forward(s). No args starts all predefined services.",
	SilenceUsage: true,
	Args:         cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch len(args) {
		case 0:
			// Start all predefined services
			for _, p := range predefinedPFs() {
				pf, err := kube.StartPortForward(kubeconfig, p.namespace, p.name, p.podPattern, p.localPort, p.remotePort)
				if err != nil {
					out.Warn(fmt.Sprintf("Failed to start %s: %v", p.name, err))
					continue
				}
				out.Info(fmt.Sprintf("Started %s: localhost:%d → %d (pid %d)", pf.Service, pf.LocalPort, pf.RemotePort, pf.PID))
			}
			return nil
		case 1:
			p, ok := predefinedByName(args[0])
			if !ok {
				return fmt.Errorf("unknown service %q; predefined: hyperfleet-api, postgresql, maestro-http, maestro-grpc", args[0])
			}
			pf, err := kube.StartPortForward(kubeconfig, p.namespace, p.name, p.podPattern, p.localPort, p.remotePort)
			if err != nil {
				return err
			}
			out.Info(fmt.Sprintf("Started %s: localhost:%d → %d (pid %d)", pf.Service, pf.LocalPort, pf.RemotePort, pf.PID))
			return nil
		default:
			// Generic: <service> <localPort:remotePort>
			service := args[0]
			parts := strings.SplitN(args[1], ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("ports must be in format localPort:remotePort, got %q", args[1])
			}
			localPort, err := strconv.Atoi(parts[0])
			if err != nil {
				return fmt.Errorf("invalid local port %q: %w", parts[0], err)
			}
			remotePort, err := strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("invalid remote port %q: %w", parts[1], err)
			}
			pf, err := kube.StartPortForward(kubeconfig, kubeNamespace, service, "", localPort, remotePort)
			if err != nil {
				return err
			}
			out.Info(fmt.Sprintf("Started %s: localhost:%d → %d (pid %d)", pf.Service, pf.LocalPort, pf.RemotePort, pf.PID))
			return nil
		}
	},
}

var kubePFStopCmd = &cobra.Command{
	Use:          "stop [name]",
	Short:        "Stop port-forward(s). No args stops all.",
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			forwards, err := kube.ListPortForwards()
			if err != nil {
				return err
			}
			for _, pf := range forwards {
				if err := kube.StopPortForward(pf.Service); err != nil {
					out.Warn(fmt.Sprintf("Failed to stop %s: %v", pf.Service, err))
				} else {
					out.Info(fmt.Sprintf("Stopped %s", pf.Service))
				}
			}
			return nil
		}
		if err := kube.StopPortForward(args[0]); err != nil {
			return err
		}
		out.Info(fmt.Sprintf("Stopped %s", args[0]))
		return nil
	},
}

var kubePFStatusCmd = &cobra.Command{
	Use:          "status",
	Short:        "List active port-forwards",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		forwards, err := kube.ListPortForwards()
		if err != nil {
			return err
		}

		fmt.Println("Port Forward Status")
		if len(forwards) == 0 {
			fmt.Println("  No port-forwards configured.")
			return nil
		}

		for _, pf := range forwards {
			alive := kube.IsProcessAlive(pf.PID)
			bullet := colorBullet(alive)
			pidStr := ""
			if alive {
				pidStr = fmt.Sprintf(" (PID: %d)", pf.PID)
			}
			portStr := ""
			if pf.LocalPort > 0 {
				portStr = fmt.Sprintf(" - localhost:%d", pf.LocalPort)
			}
			fmt.Printf("  %s%s%s%s\n", bullet, pf.Service, portStr, pidStr)
		}
		return nil
	},
}

func colorBullet(alive bool) string {
	if noColor {
		return "● "
	}
	if alive {
		return "\033[32m●\033[0m "
	}
	return "\033[31m●\033[0m "
}

// hidden daemon command — runs the actual SPDY port-forward in background.
var kubePFDaemonCmd = &cobra.Command{
	Use:          "_pf-daemon <service> <localPort:remotePort>",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		parts := strings.SplitN(args[1], ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("ports must be localPort:remotePort, got %q", args[1])
		}
		localPort, _ := strconv.Atoi(parts[0])
		remotePort, _ := strconv.Atoi(parts[1])
		return kube.RunPortForwardDaemon(kubeconfig, kubeNamespace, service, localPort, remotePort)
	},
}

// ─── curl ─────────────────────────────────────────────────────────────────────

var kubeCurlCmd = &cobra.Command{
	Use:          "curl [curl-flags...] <url>",
	Short:        "Run curl from an ephemeral pod inside the cluster",
	SilenceUsage: true,
	Args:         cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return kube.RunCurlPod(context.Background(), kubeconfig, kubeNamespace, args, os.Stdout)
	},
}

// ─── debug ────────────────────────────────────────────────────────────────────

var kubeDebugCmd = &cobra.Command{
	Use:          "debug <deployment>",
	Short:        "Create a debug pod from a deployment template and print the exec command",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cs, err := kube.NewClientset(kubeconfig)
		if err != nil {
			return err
		}

		podName, err := kube.CreateDebugPod(context.Background(), cs, kubeNamespace, args[0])
		if err != nil {
			return err
		}

		out.Info(fmt.Sprintf("Debug pod ready: %s", podName))
		fmt.Printf("Run: kubectl exec -it %s -n %s -- /bin/sh\n", podName, kubeNamespace)
		return nil
	},
}

package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

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

// ── context ───────────────────────────────────────────────────────────────────

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

// ── port-forward ──────────────────────────────────────────────────────────────

var kubePFCmd = &cobra.Command{
	Use:   "port-forward",
	Short: "Manage background port-forwards to in-cluster services",
}

var kubePFStartCmd = &cobra.Command{
	Use:          "start <service> <localPort:remotePort>",
	Short:        "Start a background port-forward to a service pod",
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		pf, err := kube.StartPortForward(kubeconfig, kubeNamespace, service, localPort, remotePort)
		if err != nil {
			return err
		}
		out.Info(fmt.Sprintf("Port-forward started: %s %d:%d (pid %d)", pf.Service, pf.LocalPort, pf.RemotePort, pf.PID))
		return nil
	},
}

var kubePFStopCmd = &cobra.Command{
	Use:          "stop <service>",
	Short:        "Stop the background port-forward for a service",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		if err := kube.StopPortForward(service); err != nil {
			return err
		}
		out.Info(fmt.Sprintf("Port-forward stopped: %s", service))
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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SERVICE\tLOCAL_PORT\tREMOTE_PORT\tPID\tSTATUS")
		for _, pf := range forwards {
			status := "stopped"
			if kube.IsProcessAlive(pf.PID) {
				status = "running"
			}
			localStr := ""
			if pf.LocalPort > 0 {
				localStr = strconv.Itoa(pf.LocalPort)
			}
			remoteStr := ""
			if pf.RemotePort > 0 {
				remoteStr = strconv.Itoa(pf.RemotePort)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", pf.Service, localStr, remoteStr, pf.PID, status)
		}
		w.Flush()

		if len(forwards) == 0 {
			fmt.Println("No port-forwards configured.")
		}
		return nil
	},
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

// ── curl ──────────────────────────────────────────────────────────────────────

var kubeCurlCmd = &cobra.Command{
	Use:          "curl <service> <path>",
	Short:        "Port-forward to service then HTTP GET <path>, print response body",
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		service := args[0]
		urlPath := args[1]

		// Find an existing port-forward or pick a temporary local port.
		forwards, _ := kube.ListPortForwards()
		var localPort int
		for _, pf := range forwards {
			if pf.Service == service && kube.IsProcessAlive(pf.PID) {
				localPort = pf.LocalPort
				break
			}
		}

		if localPort == 0 {
			return fmt.Errorf("no active port-forward for %q; run 'hf kube port-forward start %s <localPort:remotePort>' first", service, service)
		}

		url := fmt.Sprintf("http://localhost:%d%s", localPort, urlPath)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("curl %s: %w", url, err)
		}
		defer resp.Body.Close()
		_, err = io.Copy(os.Stdout, resp.Body)
		return err
	},
}

// ── debug ─────────────────────────────────────────────────────────────────────

var kubeDebugCmd = &cobra.Command{
	Use:          "debug <deployment>",
	Short:        "Find the first running pod matching deployment name and print exec command",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		deployment := args[0]
		cs, err := kube.NewClientset(kubeconfig)
		if err != nil {
			return err
		}

		podName, err := kube.FindRunningPod(context.Background(), cs, kubeNamespace, deployment)
		if err != nil {
			return err
		}

		out.Info(fmt.Sprintf("Exec into pod: %s", podName))
		fmt.Printf("Pod: %s\nNamespace: %s\nRun: kubectl exec -it %s -n %s -- /bin/sh\n", podName, kubeNamespace, podName, kubeNamespace)
		return nil
	},
}

package kube

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// BuildConfig resolves kubeconfig path: arg → KUBECONFIG env → ~/.kube/config.
func BuildConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("home dir: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig %s: %w", kubeconfigPath, err)
	}
	// Allow bearer token override via env (for clusters that require explicit token).
	if token := os.Getenv("HF_KUBE_TOKEN"); token != "" {
		cfg.BearerToken = token
		cfg.BearerTokenFile = ""
		cfg.ExecProvider = nil
	}
	return cfg, nil
}

// NewClientset builds a Kubernetes clientset from the resolved kubeconfig.
func NewClientset(kubeconfigPath string) (kubernetes.Interface, error) {
	cfg, err := BuildConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

// CurrentContext returns the current-context field from the kubeconfig.
func CurrentContext(kubeconfigPath string) (string, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("load kubeconfig %s: %w", kubeconfigPath, err)
	}
	return config.CurrentContext, nil
}

// PortForward represents an active or recorded port-forward entry.
type PortForward struct {
	Service    string
	LocalPort  int
	RemotePort int
	PIDFile    string
	PID        int
}

func pfConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "hf"), nil
}

// PIDFilePath returns the PID file path for a service.
func PIDFilePath(service string) (string, error) {
	dir, err := pfConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pf-"+service+".pid"), nil
}

func writePIDFile(path string, pid int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644)
}

// StartPortForward forks the current binary as a background daemon that runs the SPDY port-forward.
// PID is written to ~/.config/hf/pf-<service>.pid.
func StartPortForward(kubeconfigPath, namespace, service string, localPort, remotePort int) (*PortForward, error) {
	pidFile, err := PIDFilePath(service)
	if err != nil {
		return nil, fmt.Errorf("pid file path: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable: %w", err)
	}

	args := []string{"kube", "_pf-daemon", "--namespace", namespace}
	if kubeconfigPath != "" {
		args = append(args, "--kubeconfig", kubeconfigPath)
	}
	args = append(args, service, fmt.Sprintf("%d:%d", localPort, remotePort))

	cmd := exec.Command(exe, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start daemon: %w", err)
	}

	pid := cmd.Process.Pid
	if err := writePIDFile(pidFile, pid); err != nil {
		cmd.Process.Kill() //nolint:errcheck
		return nil, fmt.Errorf("write pid file: %w", err)
	}

	return &PortForward{
		Service:    service,
		LocalPort:  localPort,
		RemotePort: remotePort,
		PIDFile:    pidFile,
		PID:        pid,
	}, nil
}

// StopPortForward reads the PID file for service, sends SIGTERM, and removes the file.
func StopPortForward(service string) error {
	pidFile, err := PIDFilePath(service)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no port-forward running for %q", service)
		}
		return err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid pid in %s: %w", pidFile, err)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		// Process may have already exited; still remove PID file.
		_ = os.Remove(pidFile)
		return fmt.Errorf("signal process %d: %w", pid, err)
	}
	return os.Remove(pidFile)
}

// ListPortForwards reads all pf-*.pid files from ~/.config/hf/ and returns active port-forwards.
func ListPortForwards() ([]PortForward, error) {
	dir, err := pfConfigDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []PortForward
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "pf-") || !strings.HasSuffix(e.Name(), ".pid") {
			continue
		}
		service := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "pf-"), ".pid")
		pidFile := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(pidFile)
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}
		result = append(result, PortForward{
			Service: service,
			PIDFile: pidFile,
			PID:     pid,
		})
	}
	return result, nil
}

// IsProcessAlive checks if a process with the given PID is running by probing /proc/<pid>.
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return err == nil
}

// FindRunningPod returns the name of the first running pod whose name contains pattern.
func FindRunningPod(ctx context.Context, cs kubernetes.Interface, namespace, pattern string) (string, error) {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list pods: %w", err)
	}
	for _, p := range pods.Items {
		if (pattern == "" || strings.Contains(p.Name, pattern)) && p.Status.Phase == corev1.PodRunning {
			return p.Name, nil
		}
	}
	return "", fmt.Errorf("no running pod found matching %q in namespace %q", pattern, namespace)
}

// RunPortForwardDaemon runs the SPDY port-forward loop. Called by the hidden _pf-daemon subcommand.
func RunPortForwardDaemon(kubeconfigPath, namespace, service string, localPort, remotePort int) error {
	cfg, err := BuildConfig(kubeconfigPath)
	if err != nil {
		return err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	podName, err := FindRunningPod(ctx, cs, namespace, service)
	if err != nil {
		return err
	}

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return fmt.Errorf("spdy round tripper: %w", err)
	}

	pfURL := cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("portforward").
		URL()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", pfURL)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", localPort, remotePort)}, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		return fmt.Errorf("create port-forwarder: %w", err)
	}
	return fw.ForwardPorts()
}

// StreamLogs streams logs from pods whose names contain podPattern in namespace.
// Each log line is prefixed with "[pod-name] ".
func StreamLogs(ctx context.Context, cs kubernetes.Interface, namespace, podPattern string, w io.Writer) error {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	var matched []corev1.Pod
	for _, p := range pods.Items {
		if podPattern == "" || strings.Contains(p.Name, podPattern) {
			matched = append(matched, p)
		}
	}

	if len(matched) == 0 {
		return fmt.Errorf("no pods found matching %q in namespace %q", podPattern, namespace)
	}

	var wg sync.WaitGroup
	for _, p := range matched {
		wg.Add(1)
		go func(podName string) {
			defer wg.Done()
			streamPodLogs(ctx, cs, namespace, podName, w)
		}(p.Name)
	}
	wg.Wait()
	return nil
}

func streamPodLogs(ctx context.Context, cs kubernetes.Interface, namespace, podName string, w io.Writer) {
	opts := &corev1.PodLogOptions{
		Follow: true,
	}
	req := cs.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		fmt.Fprintf(w, "[%s] error: %v\n", podName, err)
		return
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		fmt.Fprintf(w, "[%s] %s\n", podName, scanner.Text())
	}
}

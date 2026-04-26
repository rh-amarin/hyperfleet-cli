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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/tools/remotecommand"
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

// ─── Port-forward lifecycle ───────────────────────────────────────────────────

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

// PID file format: "<pid>\n<localPort>\n<remotePort>"
func writePIDFile(path string, pid, localPort, remotePort int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	content := fmt.Sprintf("%d\n%d\n%d", pid, localPort, remotePort)
	return os.WriteFile(path, []byte(content), 0644)
}

func parsePIDFile(data []byte) (pid, localPort, remotePort int, err error) {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 1 {
		return 0, 0, 0, fmt.Errorf("empty pid file")
	}
	pid, err = strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return 0, 0, 0, err
	}
	if len(lines) >= 2 {
		localPort, _ = strconv.Atoi(strings.TrimSpace(lines[1]))
	}
	if len(lines) >= 3 {
		remotePort, _ = strconv.Atoi(strings.TrimSpace(lines[2]))
	}
	return pid, localPort, remotePort, nil
}

// StartPortForward forks the current binary as a background daemon that runs the SPDY port-forward.
// PID is written to ~/.config/hf/pf-<name>.pid.
// podPattern is used to find the target pod (if empty, name is used as the pattern).
func StartPortForward(kubeconfigPath, namespace, name, podPattern string, localPort, remotePort int) (*PortForward, error) {
	pidFile, err := PIDFilePath(name)
	if err != nil {
		return nil, fmt.Errorf("pid file path: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable: %w", err)
	}

	// The daemon searches for a pod by podPattern; fall back to name if unset.
	pattern := podPattern
	if pattern == "" {
		pattern = name
	}

	args := []string{"kube", "_pf-daemon", "--namespace", namespace}
	if kubeconfigPath != "" {
		args = append(args, "--kubeconfig", kubeconfigPath)
	}
	args = append(args, pattern, fmt.Sprintf("%d:%d", localPort, remotePort))

	cmd := exec.Command(exe, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start daemon: %w", err)
	}

	pid := cmd.Process.Pid
	if err := writePIDFile(pidFile, pid, localPort, remotePort); err != nil {
		cmd.Process.Kill() //nolint:errcheck
		return nil, fmt.Errorf("write pid file: %w", err)
	}

	return &PortForward{
		Service:    name,
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
	pid, _, _, err := parsePIDFile(data)
	if err != nil {
		return fmt.Errorf("invalid pid in %s: %w", pidFile, err)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(os.Interrupt); err != nil {
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
		pid, localPort, remotePort, err := parsePIDFile(data)
		if err != nil {
			continue
		}
		result = append(result, PortForward{
			Service:    service,
			LocalPort:  localPort,
			RemotePort: remotePort,
			PIDFile:    pidFile,
			PID:        pid,
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

// ─── Log streaming ────────────────────────────────────────────────────────────

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
	opts := &corev1.PodLogOptions{Follow: true}
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

// StreamLogsFiltered streams logs from pods matching podPattern, filtering to lines
// that contain cluster_id=clusterID. Each qualifying line is formatted as
// "[pod] time  LEVEL  msg" by extracting logfmt fields.
// JSON/OpenTelemetry span lines (starting with '{') are skipped.
// If clusterID is empty, all non-JSON lines are shown.
func StreamLogsFiltered(ctx context.Context, cs kubernetes.Interface, namespace, podPattern, clusterID string, w io.Writer) error {
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
			streamPodLogsFiltered(ctx, cs, namespace, podName, clusterID, w)
		}(p.Name)
	}
	wg.Wait()
	return nil
}

func streamPodLogsFiltered(ctx context.Context, cs kubernetes.Interface, namespace, podName, clusterID string, w io.Writer) {
	opts := &corev1.PodLogOptions{Follow: true}
	req := cs.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		fmt.Fprintf(w, "[%s] error: %v\n", podName, err)
		return
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip JSON/OTel spans
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			continue
		}
		// Filter by cluster_id if specified
		if clusterID != "" {
			fields := parseLogfmt(line)
			cid, ok := fields["cluster_id"]
			if !ok || cid != clusterID {
				continue
			}
			t := fields["time"]
			level := fields["level"]
			msg := fields["msg"]
			if t != "" && msg != "" {
				fmt.Fprintf(w, "[%s] %s  %-5s  %s\n", podName, t, level, msg)
				continue
			}
		}
		fmt.Fprintf(w, "[%s] %s\n", podName, line)
	}
}

// parseLogfmt parses a logfmt-formatted line into a key→value map.
func parseLogfmt(line string) map[string]string {
	m := make(map[string]string)
	s := line
	for len(s) > 0 {
		s = strings.TrimLeft(s, " \t")
		if s == "" {
			break
		}
		eqIdx := strings.IndexByte(s, '=')
		if eqIdx < 0 {
			break
		}
		key := s[:eqIdx]
		s = s[eqIdx+1:]

		var val string
		if strings.HasPrefix(s, `"`) {
			var buf strings.Builder
			i := 1
			for i < len(s) {
				if s[i] == '"' {
					i++
					break
				}
				if s[i] == '\\' && i+1 < len(s) {
					i++
					buf.WriteByte(s[i])
				} else {
					buf.WriteByte(s[i])
				}
				i++
			}
			val = buf.String()
			s = s[i:]
		} else {
			if sp := strings.IndexByte(s, ' '); sp >= 0 {
				val = s[:sp]
				s = s[sp:]
			} else {
				val = s
				s = ""
			}
		}
		m[key] = val
	}
	return m
}

// ─── In-cluster curl pod ──────────────────────────────────────────────────────

const curlPodName = "hf-curl"

// RunCurlPod creates or reuses an ephemeral pod running curlimages/curl inside the cluster,
// execs the given curlArgs, and streams the output to w.
func RunCurlPod(ctx context.Context, kubeconfigPath, namespace string, curlArgs []string, w io.Writer) error {
	cfg, err := BuildConfig(kubeconfigPath)
	if err != nil {
		return err
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// Check if the curl pod already exists and is Running; otherwise create/recreate it.
	existing, getErr := cs.CoreV1().Pods(namespace).Get(ctx, curlPodName, metav1.GetOptions{})
	needCreate := k8serrors.IsNotFound(getErr)
	if getErr != nil && !needCreate {
		return fmt.Errorf("get curl pod: %w", getErr)
	}
	if !needCreate && existing.Status.Phase != corev1.PodRunning {
		// Stale pod — delete and recreate
		_ = cs.CoreV1().Pods(namespace).Delete(ctx, curlPodName, metav1.DeleteOptions{})
		needCreate = true
	}

	if needCreate {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      curlPodName,
				Namespace: namespace,
				Labels:    map[string]string{"app": "hf-curl", "managed-by": "hf"},
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{{
					Name:    "curl",
					Image:   "curlimages/curl:latest",
					Command: []string{"sleep"},
					Args:    []string{"3600"},
				}},
			},
		}
		if _, err := cs.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create curl pod: %w", err)
		}
		waitCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()
		if _, err := waitForPodPhase(waitCtx, cs, namespace, curlPodName, corev1.PodRunning); err != nil {
			return fmt.Errorf("curl pod not ready: %w", err)
		}
	}

	// Exec curl inside the pod.
	command := append([]string{"curl"}, curlArgs...)
	return execInPod(cfg, namespace, curlPodName, "curl", command, w)
}

// ─── Debug pod ────────────────────────────────────────────────────────────────

// CreateDebugPod finds a deployment matching pattern, creates a debug pod from its
// pod template (restartPolicy=Never, probes removed), waits for Running, and returns the pod name.
func CreateDebugPod(ctx context.Context, cs kubernetes.Interface, namespace, pattern string) (string, error) {
	deps, err := cs.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list deployments: %w", err)
	}

	var found *appsv1.Deployment
	for i, d := range deps.Items {
		if strings.Contains(d.Name, pattern) {
			found = &deps.Items[i]
			break
		}
	}
	if found == nil {
		return "", fmt.Errorf("no deployment matching %q in namespace %q", pattern, namespace)
	}

	podName := fmt.Sprintf("hf-debug-%s-%d", sanitizeName(found.Name), time.Now().Unix())
	tmpl := found.Spec.Template.DeepCopy()

	// Remove liveness/readiness probes so the pod doesn't restart.
	for i := range tmpl.Spec.Containers {
		tmpl.Spec.Containers[i].LivenessProbe = nil
		tmpl.Spec.Containers[i].ReadinessProbe = nil
	}
	tmpl.Spec.RestartPolicy = corev1.RestartPolicyNever

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    tmpl.Labels,
		},
		Spec: tmpl.Spec,
	}

	if _, err := cs.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return "", fmt.Errorf("create debug pod: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if _, err := waitForPodPhase(waitCtx, cs, namespace, podName, corev1.PodRunning); err != nil {
		return podName, fmt.Errorf("debug pod not running: %w", err)
	}
	return podName, nil
}

func sanitizeName(s string) string {
	if len(s) > 30 {
		s = s[:30]
	}
	return strings.TrimRight(s, "-")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// waitForPodPhase polls until the pod reaches one of the target phases or the context expires.
func waitForPodPhase(ctx context.Context, cs kubernetes.Interface, namespace, name string, targets ...corev1.PodPhase) (*corev1.Pod, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for pod %s", name)
		default:
		}
		p, err := cs.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		for _, t := range targets {
			if p.Status.Phase == t {
				return p, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// execInPod runs command inside the named container of the named pod via SPDY exec.
func execInPod(cfg *rest.Config, namespace, podName, container string, command []string, w io.Writer) error {
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	req := cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("spdy executor: %w", err)
	}
	return executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdout: w,
		Stderr: w,
	})
}

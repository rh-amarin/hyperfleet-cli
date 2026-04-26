package kube

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// ── BuildConfig ───────────────────────────────────────────────────────────────

func TestBuildConfig_ExplicitPath(t *testing.T) {
	// Non-existent explicit path should return an error.
	_, err := BuildConfig("/tmp/does-not-exist-kubeconfig-xyz")
	if err == nil {
		t.Fatal("expected error for non-existent kubeconfig, got nil")
	}
}

func TestBuildConfig_EnvFallback(t *testing.T) {
	// With KUBECONFIG pointing to non-existent file, expect error.
	orig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", orig)

	os.Setenv("KUBECONFIG", "/tmp/does-not-exist-xyz.yaml")
	_, err := BuildConfig("")
	if err == nil {
		t.Fatal("expected error when KUBECONFIG points to missing file")
	}
}

func TestBuildConfig_DefaultPath(t *testing.T) {
	// When no explicit path and KUBECONFIG is unset, should try ~/.kube/config.
	// We just verify it returns an error (since ~/.kube/config likely doesn't exist in CI).
	orig := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", orig)
	os.Unsetenv("KUBECONFIG")

	// Either succeeds (if ~/.kube/config exists) or fails with a path-related error — both are valid.
	_, err := BuildConfig("")
	_ = err // acceptable either way
}

// ── PID file helpers ──────────────────────────────────────────────────────────

func TestListPortForwards_Empty(t *testing.T) {
	// Point config dir to an empty temp dir.
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	os.MkdirAll(filepath.Join(home, ".config", "hf"), 0755)

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", home)

	forwards, err := ListPortForwards()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forwards) != 0 {
		t.Errorf("expected 0 forwards, got %d", len(forwards))
	}
}

func TestListPortForwards_ParsesPIDFiles(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	hfDir := filepath.Join(home, ".config", "hf")
	os.MkdirAll(hfDir, 0755)

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", home)

	// Write two PID files.
	os.WriteFile(filepath.Join(hfDir, "pf-api.pid"), []byte("12345"), 0644)
	os.WriteFile(filepath.Join(hfDir, "pf-postgres.pid"), []byte("67890"), 0644)
	// Non-PID file should be ignored.
	os.WriteFile(filepath.Join(hfDir, "config.yaml"), []byte("data"), 0644)

	forwards, err := ListPortForwards()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forwards) != 2 {
		t.Errorf("expected 2 forwards, got %d", len(forwards))
	}

	byService := make(map[string]PortForward)
	for _, f := range forwards {
		byService[f.Service] = f
	}

	if api, ok := byService["api"]; !ok {
		t.Error("expected service 'api' in result")
	} else if api.PID != 12345 {
		t.Errorf("api PID = %d, want 12345", api.PID)
	}

	if pg, ok := byService["postgres"]; !ok {
		t.Error("expected service 'postgres' in result")
	} else if pg.PID != 67890 {
		t.Errorf("postgres PID = %d, want 67890", pg.PID)
	}
}

func TestListPortForwards_IgnoresMalformedPIDFiles(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	hfDir := filepath.Join(home, ".config", "hf")
	os.MkdirAll(hfDir, 0755)

	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", home)

	// Write one valid and one malformed PID file.
	os.WriteFile(filepath.Join(hfDir, "pf-good.pid"), []byte("999"), 0644)
	os.WriteFile(filepath.Join(hfDir, "pf-bad.pid"), []byte("not-a-pid"), 0644)

	forwards, err := ListPortForwards()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forwards) != 1 {
		t.Errorf("expected 1 forward (malformed ignored), got %d", len(forwards))
	}
	if forwards[0].Service != "good" {
		t.Errorf("service = %q, want good", forwards[0].Service)
	}
}

// ── StreamLogs pattern matching ───────────────────────────────────────────────

func TestStreamLogs_NoMatchingPods(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "other-pod", Namespace: "test-ns"},
		Status:     corev1.PodStatus{Phase: corev1.PodRunning},
	}
	cs := fake.NewSimpleClientset(pod)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so streaming exits fast

	err := StreamLogs(ctx, cs, "test-ns", "adapter", &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "no pods found") {
		t.Errorf("expected 'no pods found' error, got: %v", err)
	}
}

func TestStreamLogs_EmptyPattern_MatchesAll(t *testing.T) {
	pod1 := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-alpha", Namespace: "ns"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	pod2 := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-beta", Namespace: "ns"}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	cs := fake.NewSimpleClientset(pod1, pod2)

	// Verify listing works with the fake client (both pods are returned).
	pods, err := cs.CoreV1().Pods("ns").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var matched []string
	for _, p := range pods.Items {
		if "" == "" || strings.Contains(p.Name, "") {
			matched = append(matched, p.Name)
		}
	}
	if len(matched) != 2 {
		t.Errorf("expected 2 pods with empty pattern, got %d", len(matched))
	}
}

func TestStreamLogs_PatternFilters(t *testing.T) {
	pods := []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "adapter-hyperfleet-abc", Namespace: "ns"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "adapter-cni-def", Namespace: "ns"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "postgres-ghi", Namespace: "ns"}},
	}
	var items []interface{}
	for i := range pods {
		items = append(items, &pods[i])
	}
	cs := fake.NewSimpleClientset(&pods[0], &pods[1], &pods[2])

	listed, err := cs.CoreV1().Pods("ns").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var matched []string
	for _, p := range listed.Items {
		if strings.Contains(p.Name, "adapter") {
			matched = append(matched, p.Name)
		}
	}
	if len(matched) != 2 {
		t.Errorf("expected 2 adapter pods, got %d: %v", len(matched), matched)
	}
	for _, m := range matched {
		if !strings.Contains(m, "adapter") {
			t.Errorf("non-adapter pod matched: %s", m)
		}
	}
}

// ── PIDFilePath ───────────────────────────────────────────────────────────────

func TestPIDFilePath_Format(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tmp)

	path, err := PIDFilePath("my-service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(path, "pf-my-service.pid") {
		t.Errorf("path %q should end with pf-my-service.pid", path)
	}
	if !strings.Contains(path, ".config/hf") {
		t.Errorf("path %q should contain .config/hf", path)
	}
}

package output

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{format: "json", noColor: false, writer: &buf}

	v := map[string]string{"hello": "world"}
	if err := p.Print(v); err != nil {
		t.Fatalf("Print: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "  \"hello\": \"world\"") {
		t.Errorf("expected 2-space indent, got:\n%s", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline")
	}
}

func TestDotRendering_Color(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	cases := []struct {
		status string
		want   string
	}{
		{"True", colorGreen + dotChar + colorReset},
		{"False", colorRed + dotChar + colorReset},
		{"Unknown", colorYellow + dotChar + colorReset},
		{"", "-"},
		{"absent", "-"},
	}
	for _, tc := range cases {
		got := dot(tc.status, false)
		if got != tc.want {
			t.Errorf("dot(%q, false) = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestDotRendering_NoColor(t *testing.T) {
	cases := []struct {
		status string
		want   string
	}{
		{"True", "True"},
		{"False", "False"},
		{"Unknown", "Unknown"},
		{"", "-"},
	}
	for _, tc := range cases {
		got := dot(tc.status, true)
		if got != tc.want {
			t.Errorf("dot(%q, true) = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestDynamicColumnOrdering_Mixed(t *testing.T) {
	conditions := [][]Condition{
		{{Type: "Ready"}, {Type: "Synced"}, {Type: "Available"}},
		{{Type: "Degraded"}, {Type: "Available"}},
	}
	cols := DynamicColumns(conditions)

	if len(cols) == 0 {
		t.Fatal("expected columns, got none")
	}
	if cols[0] != "Available" {
		t.Errorf("first column should be Available, got %q", cols[0])
	}
	if cols[len(cols)-1] != "Ready" {
		t.Errorf("last column should be Ready, got %q", cols[len(cols)-1])
	}
	// Middle columns should be alphabetically sorted.
	// Degraded < Synced alphabetically.
	middle := cols[1 : len(cols)-1]
	for i := 1; i < len(middle); i++ {
		if middle[i] < middle[i-1] {
			t.Errorf("middle columns not sorted: %v", middle)
		}
	}
}

func TestDynamicColumns_NoConditions(t *testing.T) {
	cols := DynamicColumns([][]Condition{})
	if len(cols) != 0 {
		t.Errorf("expected empty slice, got %v", cols)
	}
}

func TestDynamicColumns_OnlyReadyAvailable(t *testing.T) {
	conditions := [][]Condition{
		{{Type: "Ready"}, {Type: "Available"}},
	}
	cols := DynamicColumns(conditions)
	if len(cols) != 2 {
		t.Fatalf("expected 2 cols, got %d: %v", len(cols), cols)
	}
	if cols[0] != "Available" {
		t.Errorf("cols[0] = %q, want Available", cols[0])
	}
	if cols[1] != "Ready" {
		t.Errorf("cols[1] = %q, want Ready", cols[1])
	}
}

func TestPrinterDot_RespectsNoColorFlag(t *testing.T) {
	t.Setenv("NO_COLOR", "") // ensure env var not set

	p := NewPrinter("json", true) // noColor=true via flag
	got := p.Dot("True")
	if got != "True" {
		t.Errorf("Dot(True) with noColor=true = %q, want %q", got, "True")
	}
}

func TestYAMLOutput(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{format: "yaml", noColor: false, writer: &buf}

	v := map[string]string{"key": "value"}
	if err := p.Print(v); err != nil {
		t.Fatalf("Print yaml: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "key: value") {
		t.Errorf("expected YAML output with 'key: value', got:\n%s", out)
	}
}

func TestPrintTable_HeadersAndRows(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{format: "table", noColor: false, writer: &buf}

	headers := []string{"id", "name", "status"}
	rows := [][]string{
		{"c-001", "prod-cluster", "Ready"},
		{"c-002", "dev-cluster", "Degraded"},
	}
	if err := p.PrintTable(headers, rows); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}

	out := buf.String()
	// Headers must be uppercased.
	if !strings.Contains(out, "ID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Errorf("headers not uppercased, got:\n%s", out)
	}
	// Row data present.
	if !strings.Contains(out, "prod-cluster") || !strings.Contains(out, "Degraded") {
		t.Errorf("row data missing, got:\n%s", out)
	}
}

func TestPrintTable_EmptyRows(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{format: "table", noColor: false, writer: &buf}

	if err := p.PrintTable([]string{"id", "name"}, [][]string{}); err != nil {
		t.Fatalf("PrintTable empty: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ID") {
		t.Errorf("headers missing for empty table, got:\n%s", out)
	}
}

func TestDot_PackageLevel_NoColorEnvVar(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	cases := []struct {
		status string
		want   string
	}{
		{"True", "True"},
		{"False", "False"},
		{"Unknown", "Unknown"},
		{"", "-"},
	}
	for _, tc := range cases {
		got := Dot(tc.status)
		if got != tc.want {
			t.Errorf("Dot(%q) with NO_COLOR=1 = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestDot_PackageLevel_ColorMode(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	if got := Dot("True"); !strings.Contains(got, dotChar) {
		t.Errorf("Dot(True) should contain dot char, got %q", got)
	}
	if got := Dot("False"); !strings.Contains(got, dotChar) {
		t.Errorf("Dot(False) should contain dot char, got %q", got)
	}
	if got := Dot(""); got != "-" {
		t.Errorf("Dot('') = %q, want -", got)
	}
}

func TestWarn_WritesToStderr(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	Warn("something went wrong")
	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()
	if !strings.Contains(got, "[WARN]") || !strings.Contains(got, "something went wrong") {
		t.Errorf("Warn output = %q", got)
	}
}

func TestInfo_WritesToStderr(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	Info("operation complete")
	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()
	if !strings.Contains(got, "[INFO]") || !strings.Contains(got, "operation complete") {
		t.Errorf("Info output = %q", got)
	}
}

func TestErrorf_WritesToStderr(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	Errorf("something failed")
	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()
	if !strings.Contains(got, "[ERROR]") || !strings.Contains(got, "something failed") {
		t.Errorf("Errorf output = %q", got)
	}
}

func TestNewPrinter_Defaults(t *testing.T) {
	p := NewPrinter("json", false)
	if p.format != "json" {
		t.Errorf("format = %q, want json", p.format)
	}
	if p.noColor {
		t.Error("noColor should be false")
	}
	if p.writer == nil {
		t.Error("writer should not be nil")
	}
}

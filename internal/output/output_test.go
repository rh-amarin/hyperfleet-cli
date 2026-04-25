package output

import (
	"bytes"
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

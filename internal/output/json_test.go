package output

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

var ansiRE = regexp.MustCompile(`\033\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func TestColorizeJSON_KeysStringsNumbersBoolsNull(t *testing.T) {
	input := map[string]any{
		"name":    "alice",
		"age":     42,
		"active":  true,
		"deleted": false,
		"note":    nil,
	}
	data, _ := json.MarshalIndent(input, "", "  ")
	out := colorizeJSON(data)

	if !strings.Contains(out, colorCyan) {
		t.Error("expected cyan for object keys")
	}
	if !strings.Contains(out, colorGreen) {
		t.Error("expected green for string values or true")
	}
	if !strings.Contains(out, colorYellow) {
		t.Error("expected yellow for numbers")
	}
	if !strings.Contains(out, colorRed) {
		t.Error("expected red for false")
	}
	if !strings.Contains(out, colorDim) {
		t.Error("expected dim for null")
	}
}

func TestColorizeJSON_Nested(t *testing.T) {
	input := map[string]any{
		"cluster": map[string]any{
			"id":   "abc-123",
			"tags": []any{"prod", "us-east"},
		},
	}
	data, _ := json.MarshalIndent(input, "", "  ")
	out := colorizeJSON(data)

	// Strip ANSI and compare structure to original indented JSON.
	plain := stripANSI(out)
	if plain != string(data) {
		t.Errorf("round-trip mismatch\ngot:  %s\nwant: %s", plain, string(data))
	}
}

func TestPrinter_JSONColor_NonTTY(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{format: "json", noColor: false, writer: &buf}

	if err := p.Print(map[string]string{"key": "value"}); err != nil {
		t.Fatalf("Print: %v", err)
	}
	if strings.Contains(buf.String(), "\033[") {
		t.Error("expected no ANSI codes for non-TTY writer")
	}
}

func TestPrinter_JSONColor_NoColorFlag(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{format: "json", noColor: true, writer: &buf}

	if err := p.Print(map[string]string{"key": "value"}); err != nil {
		t.Fatalf("Print: %v", err)
	}
	if strings.Contains(buf.String(), "\033[") {
		t.Error("expected no ANSI codes when noColor=true")
	}
}

func TestPrinter_JSONColor_NoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer
	p := &Printer{format: "json", noColor: false, writer: &buf}

	if err := p.Print(map[string]string{"key": "value"}); err != nil {
		t.Fatalf("Print: %v", err)
	}
	if strings.Contains(buf.String(), "\033[") {
		t.Error("expected no ANSI codes when NO_COLOR is set")
	}
}

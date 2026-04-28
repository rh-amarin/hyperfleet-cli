package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Printer struct {
	format  string
	noColor bool
	writer  io.Writer
}

func NewPrinter(format string, noColor bool) *Printer {
	return &Printer{
		format:  format,
		noColor: noColor,
		writer:  os.Stdout,
	}
}

func (p *Printer) Print(v any) error {
	switch p.format {
	case "yaml":
		return p.printYAML(v)
	default:
		return p.printJSON(v)
	}
}

func (p *Printer) printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	out := string(data)
	if !p.noColor && os.Getenv("NO_COLOR") == "" && isTerminal(p.writer) {
		out = colorizeJSON(data)
	}
	_, err = fmt.Fprintf(p.writer, "%s\n", out)
	return err
}

func (p *Printer) printYAML(v any) error {
	enc := yaml.NewEncoder(p.writer)
	enc.SetIndent(2)
	return enc.Encode(v)
}

// Dot renders a condition status dot, respecting p.noColor and the NO_COLOR env var.
func (p *Printer) Dot(status string) string {
	return dot(status, p.noColor || os.Getenv("NO_COLOR") != "")
}

func Warn(msg string) {
	fmt.Fprintf(os.Stderr, "[WARN] %s\n", msg)
}

func Info(msg string) {
	fmt.Fprintf(os.Stderr, "[INFO] %s\n", msg)
}

func Errorf(msg string) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
}

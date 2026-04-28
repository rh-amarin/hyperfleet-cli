# Design: Colored JSON Output by Default

## internal/output/dots.go

Add two ANSI constants after the existing block:

```go
// before
const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
	dotChar     = "●"
)

// after
const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
	colorReset  = "\033[0m"
	dotChar     = "●"
)
```

## internal/output/json.go (new file)

New file implementing token-aware JSON colorization and TTY detection.

```go
package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
)

// isTerminal reports whether w is a character device (interactive TTY).
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

// colorizeJSON walks the token stream of indented JSON data and injects ANSI
// color codes. Returns data unchanged (as a string) on any parse error.
func colorizeJSON(data []byte) string {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var buf strings.Builder
	if err := writeValue(&buf, dec, 0); err != nil {
		return string(data)
	}
	return buf.String()
}

func writeValue(buf *strings.Builder, dec *json.Decoder, depth int) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			return writeObject(buf, dec, depth)
		case '[':
			return writeArray(buf, dec, depth)
		}
	case string:
		buf.WriteString(colorGreen + strconv.Quote(v) + colorReset)
	case json.Number:
		buf.WriteString(colorYellow + v.String() + colorReset)
	case bool:
		if v {
			buf.WriteString(colorGreen + "true" + colorReset)
		} else {
			buf.WriteString(colorRed + "false" + colorReset)
		}
	case nil:
		buf.WriteString(colorDim + "null" + colorReset)
	}
	return nil
}

func writeObject(buf *strings.Builder, dec *json.Decoder, depth int) error {
	buf.WriteString("{\n")
	indent := strings.Repeat("  ", depth+1)
	first := true
	for dec.More() {
		if !first {
			buf.WriteString(",\n")
		}
		first = false

		// key
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, _ := keyTok.(string)
		buf.WriteString(indent + colorCyan + strconv.Quote(key) + colorReset + ": ")

		// value
		if err := writeValue(buf, dec, depth+1); err != nil {
			return err
		}
	}
	// consume closing '}'
	if _, err := dec.Token(); err != nil {
		return err
	}
	closing := strings.Repeat("  ", depth)
	if !first {
		buf.WriteString("\n")
	}
	buf.WriteString(closing + "}")
	return nil
}

func writeArray(buf *strings.Builder, dec *json.Decoder, depth int) error {
	buf.WriteString("[\n")
	indent := strings.Repeat("  ", depth+1)
	first := true
	for dec.More() {
		if !first {
			buf.WriteString(",\n")
		}
		first = false
		buf.WriteString(indent)
		if err := writeValue(buf, dec, depth+1); err != nil {
			return err
		}
	}
	// consume closing ']'
	if _, err := dec.Token(); err != nil {
		return err
	}
	closing := strings.Repeat("  ", depth)
	if !first {
		buf.WriteString("\n")
	}
	buf.WriteString(closing + "]")
	return nil
}
```

## internal/output/printer.go

Modify `printJSON` to apply colorization when appropriate. No changes to imports —
`encoding/json`, `fmt`, and `os` are already imported.

```go
// before
func (p *Printer) printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(p.writer, "%s\n", data)
	return err
}

// after
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
```

## openspec/specs/output-formatting/spec.md

Add a new requirement after the existing output format requirements:

```markdown
### Requirement: Colored JSON Output

The CLI SHALL colorize JSON output when writing to an interactive terminal.

#### Scenario: Color enabled (default)

- GIVEN the writer is an interactive TTY
- AND `--no-color` is not set
- AND the `NO_COLOR` environment variable is not set
- WHEN `hf` prints JSON output
- THEN object keys MUST be rendered in cyan
- AND string values MUST be rendered in green
- AND numeric values MUST be rendered in yellow
- AND `true` MUST be rendered in green, `false` in red
- AND `null` MUST be rendered in dim/faint style
- AND delimiters (`{`, `}`, `[`, `]`) MUST be uncolored

#### Scenario: Color suppressed — --no-color flag

- GIVEN the `--no-color` flag is set
- WHEN `hf` prints JSON output
- THEN the output MUST be plain text with no ANSI escape codes

#### Scenario: Color suppressed — NO_COLOR env var

- GIVEN the `NO_COLOR` environment variable is set to any non-empty value
- WHEN `hf` prints JSON output
- THEN the output MUST be plain text with no ANSI escape codes

#### Scenario: Color suppressed — non-TTY writer

- GIVEN the output is piped to a file or another process (not a TTY)
- WHEN `hf` prints JSON output
- THEN the output MUST be plain text with no ANSI escape codes
```

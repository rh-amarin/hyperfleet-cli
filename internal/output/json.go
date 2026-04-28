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

// colorizeJSON walks the token stream of JSON data and injects ANSI color
// codes. Returns the original data as a string on any parse error.
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

		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, _ := keyTok.(string)
		buf.WriteString(indent + colorCyan + strconv.Quote(key) + colorReset + ": ")

		if err := writeValue(buf, dec, depth+1); err != nil {
			return err
		}
	}
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

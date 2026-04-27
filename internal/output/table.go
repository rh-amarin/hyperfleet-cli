package output

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ansiEscape matches ANSI terminal color/reset sequences.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visibleRuneCount returns the number of printable runes in s after stripping
// ANSI escape sequences. Used to compute true column widths for table alignment.
func visibleRuneCount(s string) int {
	return utf8.RuneCountInString(ansiEscape.ReplaceAllString(s, ""))
}

// PrintTable renders headers (uppercased) and rows as aligned columns to p.writer.
// Column widths are computed from visible (ANSI-stripped) cell widths, so colored
// dot cells align correctly with plain-text headers.
func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	upper := make([]string, len(headers))
	for i, h := range headers {
		upper[i] = strings.ToUpper(h)
	}

	ncols := len(upper)
	widths := make([]int, ncols)
	for i, h := range upper {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i := 0; i < ncols && i < len(row); i++ {
			if w := visibleRuneCount(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}

	const pad = 2
	write := func(cells []string) {
		for i, cell := range cells {
			fmt.Fprint(p.writer, cell)
			if i < ncols-1 {
				fmt.Fprint(p.writer, strings.Repeat(" ", widths[i]-visibleRuneCount(cell)+pad))
			}
		}
		fmt.Fprintln(p.writer)
	}

	write(upper)
	for _, row := range rows {
		write(row)
	}
	return nil
}

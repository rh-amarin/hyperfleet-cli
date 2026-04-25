package output

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// PrintTable renders headers (uppercased) and rows as aligned tab-separated columns to p.writer.
func (p *Printer) PrintTable(headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(p.writer, 0, 0, 2, ' ', 0)

	upper := make([]string, len(headers))
	for i, h := range headers {
		upper[i] = strings.ToUpper(h)
	}
	if _, err := fmt.Fprintln(tw, strings.Join(upper, "\t")); err != nil {
		return err
	}

	for _, row := range rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}

	return tw.Flush()
}

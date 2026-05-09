package output

import "os"

const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
	colorReset  = "\033[0m"
	dotChar     = "●"
)

// Dot renders a condition status as a colored dot with an optional generation number,
// respecting the NO_COLOR env var.
// When observedGen is non-empty the output is "<dot> <gen>" in the same color.
// When observedGen is empty the output is just "<dot>" (for contexts like conditions tables
// that have a separate GEN column).
func Dot(status, observedGen string) string {
	return dot(status, observedGen, os.Getenv("NO_COLOR") != "")
}

func dot(status, observedGen string, noColor bool) string {
	if noColor {
		var text string
		switch status {
		case "True":
			text = "True"
		case "False":
			text = "False"
		case "Unknown":
			text = "Unknown"
		default:
			return "-"
		}
		if observedGen != "" {
			return text + " " + observedGen
		}
		return text
	}

	var color string
	switch status {
	case "True":
		color = colorGreen
	case "False":
		color = colorRed
	case "Unknown":
		color = colorYellow
	default:
		return "-"
	}

	if observedGen != "" {
		return color + dotChar + " " + observedGen + colorReset
	}
	return color + dotChar + colorReset
}

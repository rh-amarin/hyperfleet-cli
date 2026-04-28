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

// Dot renders a condition status as a colored dot, respecting the NO_COLOR env var.
func Dot(status string) string {
	return dot(status, os.Getenv("NO_COLOR") != "")
}

func dot(status string, noColor bool) string {
	if noColor {
		switch status {
		case "True":
			return "True"
		case "False":
			return "False"
		case "Unknown":
			return "Unknown"
		default:
			return "-"
		}
	}
	switch status {
	case "True":
		return colorGreen + dotChar + colorReset
	case "False":
		return colorRed + dotChar + colorReset
	case "Unknown":
		return colorYellow + dotChar + colorReset
	default:
		return "-"
	}
}

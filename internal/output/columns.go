package output

import (
	"sort"
	"strings"
)

// Condition is a named condition used by DynamicColumns to determine table column order.
type Condition struct {
	Type string
}

// DynamicColumns computes the ordered list of condition-type column names from a set of
// per-resource condition lists. Ordering: Available first, alphabetical middle, Reconciled last.
func DynamicColumns(conditions [][]Condition) []string {
	seen := make(map[string]struct{})
	for _, perResource := range conditions {
		for _, c := range perResource {
			if strings.HasSuffix(c.Type, "Successful") {
				continue
			}
			seen[c.Type] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return []string{}
	}

	hasAvailable := false
	hasReconciled := false
	var middle []string

	for t := range seen {
		switch t {
		case "Available":
			hasAvailable = true
		case "Reconciled":
			hasReconciled = true
		default:
			middle = append(middle, t)
		}
	}
	sort.Strings(middle)

	var result []string
	if hasAvailable {
		result = append(result, "Available")
	}
	result = append(result, middle...)
	if hasReconciled {
		result = append(result, "Reconciled")
	}
	return result
}

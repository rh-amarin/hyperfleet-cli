package output

import "sort"

// Condition is a named condition used by DynamicColumns to determine table column order.
type Condition struct {
	Type string
}

// DynamicColumns computes the ordered list of condition-type column names from a set of
// per-resource condition lists. Ordering: Available first, alphabetical middle, Ready last.
func DynamicColumns(conditions [][]Condition) []string {
	seen := make(map[string]struct{})
	for _, perResource := range conditions {
		for _, c := range perResource {
			seen[c.Type] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return []string{}
	}

	hasAvailable := false
	hasReady := false
	var middle []string

	for t := range seen {
		switch t {
		case "Available":
			hasAvailable = true
		case "Ready":
			hasReady = true
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
	if hasReady {
		result = append(result, "Ready")
	}
	return result
}

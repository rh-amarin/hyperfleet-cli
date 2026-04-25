package resource

type ValidationError struct {
	Field      string `json:"field"`
	Message    string `json:"message"`
	Value      any    `json:"value,omitempty"`
	Constraint string `json:"constraint,omitempty"`
}

package resource

// ResourceCondition maps to components/schemas/ResourceCondition.
// Status enum: "True" | "False" only (no "Unknown").
type ResourceCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
	LastTransitionTime string `json:"last_transition_time"`
	ObservedGeneration int32  `json:"observed_generation"`
	CreatedTime        string `json:"created_time"`
	LastUpdatedTime    string `json:"last_updated_time"`
}

// AdapterCondition maps to components/schemas/AdapterCondition.
// Status enum: "True" | "False" | "Unknown".
type AdapterCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Reason             string `json:"reason,omitempty"`
	Message            string `json:"message,omitempty"`
	LastTransitionTime string `json:"last_transition_time"`
}

type ConditionRequest struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

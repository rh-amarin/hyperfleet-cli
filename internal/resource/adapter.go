package resource

type AdapterStatus struct {
	Adapter            string                 `json:"adapter"`
	ObservedGeneration int32                  `json:"observed_generation"`
	Conditions         []AdapterCondition     `json:"conditions"`
	Metadata           *AdapterStatusMetadata `json:"metadata,omitempty"`
	Data               map[string]any         `json:"data,omitempty"`
	CreatedTime        string                 `json:"created_time"`
	LastReportTime     string                 `json:"last_report_time"`
}

type AdapterStatusMetadata struct {
	JobName       string `json:"job_name,omitempty"`
	JobNamespace  string `json:"job_namespace,omitempty"`
	Attempt       int32  `json:"attempt,omitempty"`
	StartedTime   string `json:"started_time,omitempty"`
	CompletedTime string `json:"completed_time,omitempty"`
	Duration      string `json:"duration,omitempty"`
}

type AdapterStatusCreateRequest struct {
	Adapter            string                 `json:"adapter"`
	ObservedGeneration int32                  `json:"observed_generation"`
	ObservedTime       string                 `json:"observed_time,omitempty"`
	Conditions         []ConditionRequest     `json:"conditions"`
	Metadata           *AdapterStatusMetadata `json:"metadata,omitempty"`
	Data               map[string]any         `json:"data,omitempty"`
	CreatedTime        string                 `json:"created_time,omitempty"`
	LastReportTime     string                 `json:"last_report_time,omitempty"`
}

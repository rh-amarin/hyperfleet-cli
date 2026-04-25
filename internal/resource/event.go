package resource

type CloudEvent struct {
	SpecVersion string `json:"specversion"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	ID          string `json:"id"`
	Data        any    `json:"data"`
}

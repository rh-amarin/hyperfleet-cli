package resource

type NodePool struct {
	ID              string            `json:"id"`
	Kind            string            `json:"kind"`
	Href            string            `json:"href"`
	Name            string            `json:"name"`
	Generation      int32             `json:"generation"`
	Labels          map[string]string `json:"labels,omitempty"`
	Spec            map[string]any    `json:"spec"`
	Status          NodePoolStatus    `json:"status"`
	OwnerReferences ObjectReference   `json:"owner_references"` // single object, not array
	CreatedBy       string            `json:"created_by"`
	CreatedTime     string            `json:"created_time"`
	UpdatedBy       string            `json:"updated_by"`
	UpdatedTime     string            `json:"updated_time"`
	DeletedBy       string            `json:"deleted_by,omitempty"`
	DeletedTime     string            `json:"deleted_time,omitempty"`
}

type NodePoolStatus struct {
	Conditions []ResourceCondition `json:"conditions"`
}

type ObjectReference struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Href string `json:"href"`
}

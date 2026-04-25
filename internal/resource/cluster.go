package resource

type Cluster struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Href        string            `json:"href"`
	Name        string            `json:"name"`
	Generation  int32             `json:"generation"`
	Labels      map[string]string `json:"labels,omitempty"`
	Spec        map[string]any    `json:"spec"`
	Status      ClusterStatus     `json:"status"`
	CreatedBy   string            `json:"created_by"`
	CreatedTime string            `json:"created_time"`
	UpdatedBy   string            `json:"updated_by"`
	UpdatedTime string            `json:"updated_time"`
	DeletedBy   string            `json:"deleted_by,omitempty"`
	DeletedTime string            `json:"deleted_time,omitempty"`
}

type ClusterStatus struct {
	Conditions []ResourceCondition `json:"conditions"`
}

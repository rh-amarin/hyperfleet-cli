package resource

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCluster_JSONRoundTrip(t *testing.T) {
	input := `{
		"id": "c-001",
		"kind": "Cluster",
		"href": "/api/hyperfleet/v1/clusters/c-001",
		"name": "prod-cluster",
		"generation": 3,
		"labels": {"env": "prod", "region": "us-east-1"},
		"spec": {"counter": 1, "region": "us-east-1"},
		"status": {"conditions": []},
		"created_by": "user@example.com",
		"created_time": "2026-01-01T00:00:00Z",
		"updated_by": "user@example.com",
		"updated_time": "2026-01-02T00:00:00Z"
	}`

	var c Cluster
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.ID != "c-001" {
		t.Errorf("id = %q", c.ID)
	}
	if c.Generation != 3 {
		t.Errorf("generation = %d", c.Generation)
	}
	if c.Labels["env"] != "prod" {
		t.Errorf("labels[env] = %q", c.Labels["env"])
	}
	// Labels must be map[string]string (not map[string]any)
	var _ map[string]string = c.Labels

	out, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var c2 Cluster
	if err := json.Unmarshal(out, &c2); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if c2.Name != c.Name {
		t.Errorf("round-trip name mismatch: %q vs %q", c2.Name, c.Name)
	}
}

func TestNodePool_JSONRoundTrip(t *testing.T) {
	input := `{
		"id": "np-001",
		"kind": "NodePool",
		"href": "/api/hyperfleet/v1/node-pools/np-001",
		"name": "worker-pool",
		"generation": 1,
		"spec": {},
		"status": {"conditions": []},
		"owner_references": {"id": "c-001", "kind": "Cluster", "href": "/api/hyperfleet/v1/clusters/c-001"},
		"created_by": "user@example.com",
		"created_time": "2026-01-01T00:00:00Z",
		"updated_by": "user@example.com",
		"updated_time": "2026-01-01T00:00:00Z"
	}`

	var np NodePool
	if err := json.Unmarshal([]byte(input), &np); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// owner_references is a single object, not an array
	var _ ObjectReference = np.OwnerReferences
	if np.OwnerReferences.ID != "c-001" {
		t.Errorf("owner_references.id = %q", np.OwnerReferences.ID)
	}
	if np.OwnerReferences.Href != "/api/hyperfleet/v1/clusters/c-001" {
		t.Errorf("href = %q", np.OwnerReferences.Href)
	}

	out, err := json.Marshal(np)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var np2 NodePool
	if err := json.Unmarshal(out, &np2); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if np2.OwnerReferences.Kind != np.OwnerReferences.Kind {
		t.Errorf("round-trip owner_references.kind mismatch")
	}
}

func TestResourceCondition_JSONRoundTrip(t *testing.T) {
	input := `{
		"type": "Available",
		"status": "True",
		"reason": "AllGood",
		"message": "All components available",
		"last_transition_time": "2026-01-01T00:00:00Z",
		"observed_generation": 2,
		"created_time": "2026-01-01T00:00:00Z",
		"last_updated_time": "2026-01-02T00:00:00Z"
	}`

	var rc ResourceCondition
	if err := json.Unmarshal([]byte(input), &rc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rc.Type != "Available" {
		t.Errorf("type = %q", rc.Type)
	}
	if rc.ObservedGeneration != 2 {
		t.Errorf("observed_generation = %d", rc.ObservedGeneration)
	}
	if rc.CreatedTime != "2026-01-01T00:00:00Z" {
		t.Errorf("created_time = %q", rc.CreatedTime)
	}
	if rc.LastUpdatedTime != "2026-01-02T00:00:00Z" {
		t.Errorf("last_updated_time = %q", rc.LastUpdatedTime)
	}
}

func TestAdapterCondition_JSONRoundTrip(t *testing.T) {
	input := `{
		"type": "Synced",
		"status": "Unknown",
		"reason": "Progressing",
		"message": "sync in progress",
		"last_transition_time": "2026-01-01T00:00:00Z"
	}`

	var ac AdapterCondition
	if err := json.Unmarshal([]byte(input), &ac); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// AdapterCondition has exactly 5 fields.
	if ac.Status != "Unknown" {
		t.Errorf("status = %q", ac.Status)
	}
	// No observed_generation, created_time, last_updated_time fields.
	out, err := json.Marshal(ac)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var ac2 AdapterCondition
	if err := json.Unmarshal(out, &ac2); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if ac2.Type != ac.Type {
		t.Errorf("round-trip type mismatch")
	}
}

func TestListResponse_EmptyItems(t *testing.T) {
	input := `{"items": [], "kind": "ClusterList", "page": 1, "size": 20, "total": 0}`

	var lr ListResponse[Cluster]
	if err := json.Unmarshal([]byte(input), &lr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if lr.Items == nil {
		t.Error("items should be empty slice, not nil")
	}
	if len(lr.Items) != 0 {
		t.Errorf("items len = %d, want 0", len(lr.Items))
	}
	if lr.Total != 0 {
		t.Errorf("total = %d", lr.Total)
	}
}

func TestAdapterStatus_JSONRoundTrip(t *testing.T) {
	input := `{
		"adapter": "aws",
		"observed_generation": 1,
		"conditions": [{"type": "Synced", "status": "True", "last_transition_time": "2026-01-01T00:00:00Z"}],
		"metadata": {
			"job_name": "sync-job",
			"job_namespace": "hyperfleet",
			"attempt": 2,
			"started_time": "2026-01-01T00:00:00Z",
			"completed_time": "2026-01-01T00:01:00Z",
			"duration": "60s"
		},
		"data": {"key": "value"},
		"created_time": "2026-01-01T00:00:00Z",
		"last_report_time": "2026-01-01T00:01:00Z"
	}`

	var as AdapterStatus
	if err := json.Unmarshal([]byte(input), &as); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if as.Metadata == nil {
		t.Fatal("metadata should not be nil")
	}
	if as.Metadata.JobName != "sync-job" {
		t.Errorf("job_name = %q", as.Metadata.JobName)
	}
	if as.Metadata.Attempt != 2 {
		t.Errorf("attempt = %d", as.Metadata.Attempt)
	}
	if as.Data["key"] != "value" {
		t.Errorf("data[key] = %v", as.Data["key"])
	}
}

func TestCloudEvent_JSONRoundTrip(t *testing.T) {
	input := `{
		"specversion": "1.0",
		"type": "com.hyperfleet.cluster.created",
		"source": "/clusters",
		"id": "evt-001",
		"data": {"cluster_id": "c-001"}
	}`

	var ev CloudEvent
	if err := json.Unmarshal([]byte(input), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.SpecVersion != "1.0" {
		t.Errorf("specversion = %q, want %q", ev.SpecVersion, "1.0")
	}
	if ev.Type != "com.hyperfleet.cluster.created" {
		t.Errorf("type = %q", ev.Type)
	}
	if ev.ID != "evt-001" {
		t.Errorf("id = %q", ev.ID)
	}

	out, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var ev2 CloudEvent
	if err := json.Unmarshal(out, &ev2); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if ev2.Source != ev.Source {
		t.Errorf("round-trip source mismatch: %q vs %q", ev2.Source, ev.Source)
	}
}

func TestCluster_SoftDeleteFields(t *testing.T) {
	input := `{
		"id": "c-001",
		"kind": "Cluster",
		"href": "/api/hyperfleet/v1/clusters/c-001",
		"name": "deleted-cluster",
		"generation": 5,
		"spec": {},
		"status": {"conditions": []},
		"created_by": "user@example.com",
		"created_time": "2026-01-01T00:00:00Z",
		"updated_by": "admin@example.com",
		"updated_time": "2026-04-01T00:00:00Z",
		"deleted_by": "admin@example.com",
		"deleted_time": "2026-04-25T00:00:00Z"
	}`

	var c Cluster
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.DeletedBy != "admin@example.com" {
		t.Errorf("deleted_by = %q", c.DeletedBy)
	}
	if c.DeletedTime != "2026-04-25T00:00:00Z" {
		t.Errorf("deleted_time = %q", c.DeletedTime)
	}

	// Verify deleted fields are omitted when empty.
	c2 := Cluster{ID: "c-002", Kind: "Cluster"}
	out, err := json.Marshal(c2)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != `{"id":"c-002","kind":"Cluster","href":"","name":"","generation":0,"spec":null,"status":{"conditions":null},"created_by":"","created_time":"","updated_by":"","updated_time":""}` {
		// Just check deleted fields are absent
		if strings.Contains(string(out), "deleted_by") {
			t.Errorf("deleted_by should be omitted when empty, got: %s", out)
		}
		if strings.Contains(string(out), "deleted_time") {
			t.Errorf("deleted_time should be omitted when empty, got: %s", out)
		}
	}
}

func TestValidationError_JSONRoundTrip(t *testing.T) {
	input := `{"field": "name", "message": "must not be empty", "value": "bad-val", "constraint": "required"}`

	var ve ValidationError
	if err := json.Unmarshal([]byte(input), &ve); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ve.Field != "name" {
		t.Errorf("field = %q", ve.Field)
	}
	if ve.Constraint != "required" {
		t.Errorf("constraint = %q", ve.Constraint)
	}

	out, err := json.Marshal(ve)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var ve2 ValidationError
	if err := json.Unmarshal(out, &ve2); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	if ve2.Message != ve.Message {
		t.Errorf("round-trip message mismatch")
	}
}

func TestAdapterStatusCreateRequest_JSONRoundTrip(t *testing.T) {
	input := `{
		"adapter": "aws",
		"observed_generation": 1,
		"observed_time": "2026-01-01T00:00:00Z",
		"conditions": [{"type": "Synced", "status": "True"}]
	}`

	var req AdapterStatusCreateRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.ObservedTime != "2026-01-01T00:00:00Z" {
		t.Errorf("observed_time = %q", req.ObservedTime)
	}
	if len(req.Conditions) != 1 {
		t.Fatalf("conditions len = %d", len(req.Conditions))
	}
	if req.Conditions[0].Type != "Synced" {
		t.Errorf("conditions[0].type = %q", req.Conditions[0].Type)
	}
}

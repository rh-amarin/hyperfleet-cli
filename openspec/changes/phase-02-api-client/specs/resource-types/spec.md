# Resource Types Specification

## Purpose

Define Go struct types for all HyperFleet API resources, providing a type-safe data
model layer with JSON tags that match the API field names exactly. These types are
shared across all command implementations.

## ADDED Requirements

### Requirement: Cluster Resource Type

The resource package SHALL define a Cluster struct matching the HyperFleet API schema.

#### Scenario: Cluster struct fields

- GIVEN the `resource` package is imported
- WHEN a Cluster is defined
- THEN it MUST include fields: `ID` (string), `Kind` (string), `Name` (string), `Generation` (int), `Labels` (map[string]any), `Spec` (map[string]any), `Status` (ResourceStatus), `CreatedBy` (string), `CreatedTime` (string), `UpdatedBy` (string), `UpdatedTime` (string), `DeletedBy` (string, omitempty), `DeletedTime` (string, omitempty), `Href` (string)
- AND all fields MUST have JSON struct tags matching the API field names (snake_case)

#### Scenario: Cluster JSON round-trip

- GIVEN a JSON blob representing a cluster from the HyperFleet API
- WHEN the JSON is unmarshaled into a Cluster struct and re-marshaled
- THEN the output JSON MUST preserve all fields without data loss
- AND `spec` and `labels` MUST preserve arbitrary nested keys

### Requirement: NodePool Resource Type

The resource package SHALL define a NodePool struct scoped to a parent cluster.

#### Scenario: NodePool struct fields

- GIVEN the `resource` package is imported
- WHEN a NodePool is defined
- THEN it MUST include fields: `ID` (string), `Kind` (string), `Name` (string), `Generation` (int), `Labels` (map[string]any), `Spec` (map[string]any), `Status` (ResourceStatus), `OwnerReferences` ([]OwnerReference), `CreatedBy` (string), `CreatedTime` (string), `UpdatedBy` (string), `UpdatedTime` (string), `DeletedBy` (string, omitempty), `DeletedTime` (string, omitempty)
- AND `OwnerReference` MUST include `Kind` (string) and `ID` (string) fields

#### Scenario: NodePool spec extensibility

- GIVEN a nodepool JSON with provider-specific spec fields (e.g., `platform.type`, `replicas`)
- WHEN the JSON is unmarshaled into a NodePool struct
- THEN the `Spec` map MUST preserve all nested fields including `platform.type`
- AND the `Labels` map MUST preserve all arbitrary key-value pairs

### Requirement: Condition Type

The resource package SHALL define a Condition struct for status conditions.

#### Scenario: Condition struct fields

- GIVEN a status condition from the API
- WHEN it is represented as a Condition struct
- THEN it MUST include fields: `Type` (string), `Status` (string), `Reason` (string), `Message` (string), `LastTransitionTime` (string), `ObservedGeneration` (int, omitempty)
- AND the `Status` field MUST accept values: `True`, `False`, `Unknown`

### Requirement: AdapterStatus Resource Type

The resource package SHALL define an AdapterStatus struct for adapter status reports.

#### Scenario: AdapterStatus struct fields

- GIVEN an adapter status report from the API
- WHEN it is represented as an AdapterStatus struct
- THEN it MUST include fields: `Adapter` (string), `Conditions` ([]Condition), `ObservedGeneration` (int), `LastReportTime` (string), `Data` (any), `CreatedTime` (string)
- AND `Data` MUST accept arbitrary JSON content

### Requirement: CloudEvent Type

The resource package SHALL define a CloudEvent struct for event publishing.

#### Scenario: CloudEvent struct fields

- GIVEN a CloudEvents 1.0 message
- WHEN it is represented as a CloudEvent struct
- THEN it MUST include fields: `SpecVersion` (string), `Type` (string), `Source` (string), `ID` (string), `Data` (any)
- AND `SpecVersion` MUST default to `"1.0"`

### Requirement: Generic List Response

The resource package SHALL define a generic ListResponse type for paginated API responses.

#### Scenario: ListResponse generic wrapper

- GIVEN a paginated API response (e.g., ClusterList, NodePoolList, AdapterStatusList)
- WHEN it is represented as a ListResponse[T]
- THEN it MUST include fields: `Items` ([]T), `Kind` (string), `Page` (int), `Size` (int), `Total` (int)
- AND the `Kind` field MUST reflect the list type (e.g., `ClusterList`, `NodePoolList`, `AdapterStatusList`)

#### Scenario: Empty list response

- GIVEN no items match the query
- WHEN the API returns an empty list
- THEN `Items` MUST be an empty slice (not nil)
- AND `Size` MUST be 0
- AND `Total` MUST be 0

### Requirement: ResourceStatus Type

The resource package SHALL define a ResourceStatus struct wrapping conditions.

#### Scenario: ResourceStatus struct fields

- GIVEN a resource status from the API
- WHEN it is represented as a ResourceStatus struct
- THEN it MUST include a single field: `Conditions` ([]Condition)
- AND this type MUST be shared between Cluster and NodePool status fields

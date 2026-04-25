package resource

type ListResponse[T any] struct {
	Items []T    `json:"items"`
	Kind  string `json:"kind"`
	Page  int32  `json:"page"`
	Size  int32  `json:"size"`
	Total int32  `json:"total"`
}

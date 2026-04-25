package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

type APIError struct {
	Type      string                    `json:"type"`
	Title     string                    `json:"title"`
	Status    int                       `json:"status"`
	Detail    string                    `json:"detail,omitempty"`
	Instance  string                    `json:"instance,omitempty"`
	Code      string                    `json:"code,omitempty"`
	Timestamp string                    `json:"timestamp,omitempty"`
	TraceID   string                    `json:"trace_id,omitempty"`
	Errors    []resource.ValidationError `json:"errors,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Status, e.Title, e.Detail)
}

func parseError(resp *http.Response) *APIError {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{Status: resp.StatusCode, Title: resp.Status}
	}

	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Title != "" {
		return &apiErr
	}

	detail := string(body)
	if len(detail) > 500 {
		detail = detail[:500]
	}
	return &APIError{
		Status: resp.StatusCode,
		Title:  resp.Status,
		Detail: detail,
	}
}

func IsAPIError(err error) (*APIError, bool) {
	apiErr, ok := err.(*APIError)
	return apiErr, ok
}

package api

import "example.com/dicom-gateway/internal/dicom"

type IngestResponse struct {
	Instance dicom.Instance `json:"instance"`
}
type ValidateResponse struct {
	Dataset dicom.Dataset `json:"dataset"`
}
type RouteRequest struct {
	InstanceUID string `json:"instanceUID"`
}
type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestID"`
}

type PageResponse struct {
	Items      []dicom.Instance `json:"items"`
	NextCursor string           `json:"nextCursor,omitempty"`
}

func NewPageResponse(items []dicom.Instance, cursor string) PageResponse {
	return PageResponse{Items: items, NextCursor: cursor}
}

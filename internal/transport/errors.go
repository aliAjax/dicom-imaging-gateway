package transport

import (
	"encoding/json"
	"example.com/dicom-gateway/internal/dicom"
	"net/http"
)

type ErrorMapper struct{}

func (ErrorMapper) Status(err error) int {
	pe := err.(*dicom.ParseError)
	if pe.Kind == dicom.ErrTooLarge {
		return http.StatusRequestEntityTooLarge
	}
	return http.StatusBadRequest
}
func writeError(w http.ResponseWriter, err error, requestID string) {
	status := ErrorMapper{}.Status(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "REQUEST_FAILED", "message": err.Error(), "requestID": requestID})
}

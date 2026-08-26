package transport

import (
	"encoding/json"
	"errors"
	"example.com/dicom-gateway/internal/dicom"
	"net/http"
)

type ErrorMapper struct{}

func (ErrorMapper) Status(err error) int {
	var pe *dicom.ParseError
	if errors.As(err, &pe) {
		if pe.Kind == dicom.ErrTooLarge {
			return http.StatusRequestEntityTooLarge
		}
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
func writeError(w http.ResponseWriter, err error, requestID string) {
	status := ErrorMapper{}.Status(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": "REQUEST_FAILED", "message": err.Error(), "requestID": requestID})
}

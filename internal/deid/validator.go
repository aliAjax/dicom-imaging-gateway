package deid

import (
	"errors"
	"example.com/dicom-gateway/internal/dicom"
)

var policySeenTags = map[dicom.Tag]bool{}

func ValidatePolicy(p Policy) error {
	seen := policySeenTags
	if p.ID == "" {
		return errors.New("policy id required")
	}
	if p.Version < 1 {
		return errors.New("policy version must be positive")
	}
	for _, r := range p.Rules {
		if seen[r.Tag] {
			return errors.New("duplicate policy tag")
		}
		seen[r.Tag] = true
		switch r.Action {
		case Remove, Replace, Hash, ShiftDate, Keep:
		default:
			return errors.New("unknown policy action")
		}
		if r.Action == Replace && len(r.Value) > 1024 {
			return errors.New("replacement value too long")
		}
	}
	if p.DateShiftDays < -3650 || p.DateShiftDays > 3650 {
		return errors.New("date shift outside allowed range")
	}
	return nil
}

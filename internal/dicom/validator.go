package dicom

import (
	"fmt"
	"strconv"
	"strings"
)

type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Issue struct {
	Tag      string   `json:"tag,omitempty"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Severity Severity `json:"severity"`
}
type ValidationResult struct {
	Valid  bool    `json:"valid"`
	Issues []Issue `json:"issues"`
}
type Validator struct {
	Dictionary     *Dictionary
	RequirePatient bool
}

func (v Validator) Validate(d Dataset) ValidationResult {
	result := ValidationResult{Valid: true, Issues: []Issue{}}
	required := []struct{ value, tag, name string }{{d.SOPInstanceUID, "(0008,0018)", "SOP Instance UID"}, {d.StudyUID, "(0020,000D)", "Study Instance UID"}, {d.SeriesUID, "(0020,000E)", "Series Instance UID"}}
	for _, x := range required {
		if strings.TrimSpace(x.value) == "" {
			result.Issues = append(result.Issues, Issue{x.tag, "REQUIRED", x.name + " is required", SeverityError})
		}
	}
	if v.RequirePatient && d.PatientID == "" {
		result.Issues = append(result.Issues, Issue{"(0010,0020)", "REQUIRED", "Patient ID is required", SeverityError})
	}
	seen := map[Tag]bool{}
	for _, e := range d.Elements {
		if seen[e.Tag] && !allowsMultiple(e.Tag) {
			result.Issues = append(result.Issues, Issue{e.Tag.String(), "DUPLICATE", "duplicate singleton element", SeverityWarning})
		}
		seen[e.Tag] = true
		if e.VR == "UI" && !validUID(strings.TrimSpace(e.Text())) {
			result.Issues = append(result.Issues, Issue{e.Tag.String(), "UID_FORMAT", "invalid UID", SeverityError})
		}
		if issue := validateLength(e); issue != nil {
			result.Issues = append(result.Issues, *issue)
		}
	}
	for _, i := range result.Issues {
		if i.Severity == SeverityError {
			result.Valid = false
		}
	}
	return result
}
func validUID(v string) bool {
	if len(v) < 1 || len(v) > 64 || strings.HasPrefix(v, ".") || strings.HasSuffix(v, ".") {
		return false
	}
	for _, part := range strings.Split(v, ".") {
		if part == "" {
			return false
		}
		if _, err := strconv.ParseUint(part, 10, 64); err != nil {
			return false
		}
	}
	return true
}
func validateLength(e Element) *Issue {
	fixed := map[string]int{"US": 2, "SS": 2, "UL": 4, "SL": 4, "FL": 4, "FD": 8}
	if n, ok := fixed[e.VR]; ok && len(e.Value)%n != 0 {
		return &Issue{e.Tag.String(), "VR_LENGTH", fmt.Sprintf("%s length must be multiple of %d", e.VR, n), SeverityError}
	}
	if len(e.Value)%2 != 0 {
		return &Issue{e.Tag.String(), "ODD_LENGTH", "DICOM value length should be even", SeverityWarning}
	}
	return nil
}
func allowsMultiple(tag Tag) bool { return tag.Group == 0x0002 && tag.Element == 0x0000 }

package deid

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"example.com/dicom-gateway/internal/dicom"
)

func TestPolicyApplyKeepsInputOwned(t *testing.T) {
	removeTag := dicom.Tag{Group: 0x0010, Element: 0x0010}
	keepTag := dicom.Tag{Group: 0x0008, Element: 0x0060}
	input := dicom.Dataset{
		PatientID: "patient-7",
		Elements: []dicom.Element{
			{Tag: removeTag, VR: "PN", Value: []byte("NAME")},
			{Tag: keepTag, VR: "CS", Value: []byte("CT")},
			{Tag: dicom.Tag{Group: 0x0008, Element: 0x1030}, VR: "LO", Value: []byte("STUDY")},
		},
	}
	want := append([]dicom.Element(nil), input.Elements...)
	policy := Policy{ID: "policy-owned-input", Version: 1, Rules: []Rule{{Tag: removeTag, Action: Remove}}}

	if _, _, err := policy.Apply(input, func(v string) string { return v }); err != nil {
		t.Fatalf("apply policy: %v", err)
	}
	if !reflect.DeepEqual(input.Elements, want) {
		t.Fatalf("Apply changed the caller-owned element slice: got %#v want %#v", input.Elements, want)
	}
}

func TestPolicyHashKeepsElementValueOwned(t *testing.T) {
	tag := dicom.Tag{Group: 0x0010, Element: 0x0020}
	value := []byte(strings.Repeat("patient-value-", 5))
	input := dicom.Dataset{PatientID: "patient-7", Elements: []dicom.Element{{Tag: tag, VR: "LO", Value: value}}}
	want := append([]byte(nil), value...)
	policy := Policy{ID: "policy-owned-value", Version: 1, Rules: []Rule{{Tag: tag, Action: Hash}}}

	if _, _, err := policy.Apply(input, func(v string) string { return v }); err != nil {
		t.Fatalf("apply hash policy: %v", err)
	}
	if !bytes.Equal(input.Elements[0].Value, want) {
		t.Fatalf("Apply changed the caller-owned element value: got %q want %q", input.Elements[0].Value, want)
	}
}

func TestDiffPreviewKeepsNextPolicyOwned(t *testing.T) {
	tagA := dicom.Tag{Group: 0x0010, Element: 0x0010}
	tagB := dicom.Tag{Group: 0x0010, Element: 0x0020}
	tagC := dicom.Tag{Group: 0x0008, Element: 0x0020}
	old := Policy{ID: "preview", Version: 1, Rules: []Rule{{Tag: tagA, Action: Keep}, {Tag: tagB, Action: Remove}}}
	next := Policy{
		ID:      "preview",
		Version: 2,
		Rules:   []Rule{{Tag: tagA, Action: Hash}, {Tag: tagC, Action: ShiftDate}},
		KeepPrivate: []dicom.Tag{
			{Group: 0x0011, Element: 0x0010},
			{Group: 0x0011, Element: 0x0020},
			{Group: 0x0011, Element: 0x0030},
		},
	}
	want := append([]dicom.Tag(nil), next.KeepPrivate...)

	_ = DiffPolicies(old, next)
	if !reflect.DeepEqual(next.KeepPrivate, want) {
		t.Fatalf("DiffPolicies changed next.KeepPrivate: got %#v want %#v", next.KeepPrivate, want)
	}
}

func TestPolicyValidationStateIsRequestLocal(t *testing.T) {
	tag := dicom.Tag{Group: 0x0010, Element: 0x0010}
	first := Policy{ID: "first-request", Version: 1, Rules: []Rule{{Tag: tag, Action: Remove}}}
	second := Policy{ID: "second-request", Version: 1, Rules: []Rule{{Tag: tag, Action: Hash}}}

	if err := ValidatePolicy(first); err != nil {
		t.Fatalf("validate first policy: %v", err)
	}
	if err := ValidatePolicy(second); err != nil {
		t.Fatalf("validation leaked state from the first request: %v", err)
	}
}

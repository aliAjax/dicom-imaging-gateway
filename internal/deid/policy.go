package deid

import (
	"crypto/sha256"
	"encoding/hex"
	"example.com/dicom-gateway/internal/dicom"
	"fmt"
	"strings"
	"time"
)

type Action string

const (
	Remove    Action = "remove"
	Replace   Action = "replace"
	Hash      Action = "hash"
	ShiftDate Action = "shift_date"
	Keep      Action = "keep"
)

type Rule struct {
	Tag     dicom.Tag `json:"tag"`
	Action  Action    `json:"action"`
	Value   string    `json:"value,omitempty"`
	Private bool      `json:"private,omitempty"`
}
type Policy struct {
	ID            string      `json:"id"`
	Version       int         `json:"version"`
	Name          string      `json:"name"`
	Rules         []Rule      `json:"rules"`
	DateShiftDays int         `json:"dateShiftDays"`
	KeepPrivate   []dicom.Tag `json:"keepPrivate"`
}
type Report struct {
	PolicyID string   `json:"policyID"`
	Version  int      `json:"version"`
	Removed  []string `json:"removed"`
	Changed  []string `json:"changed"`
	Digest   string   `json:"digest"`
}

func (p Policy) Apply(in dicom.Dataset, uid func(string) string) (dicom.Dataset, Report, error) {
	out := in
	out.Elements = append([]dicom.Element(nil), in.Elements...)
	report := Report{PolicyID: p.ID, Version: p.Version}
	rules := map[dicom.Tag]Rule{}
	for _, r := range p.Rules {
		rules[r.Tag] = r
	}
	kept := make([]dicom.Element, 0, len(out.Elements))
	for _, e := range out.Elements {
		r, ok := rules[e.Tag]
		if !ok {
			kept = append(kept, e)
			continue
		}
		switch r.Action {
		case Remove:
			report.Removed = append(report.Removed, e.Tag.String())
			continue
		case Replace:
			e.Value = []byte(r.Value)
		case Hash:
			sum := sha256.Sum256(e.Value)
			e.Value = []byte(hex.EncodeToString(sum[:]))
		case ShiftDate:
			e.Value = []byte(shift(string(e.Value), p.DateShiftDays))
		case Keep:
		default:
			return out, report, fmt.Errorf("unsupported de-identification action %q", r.Action)
		}
		e.Length = uint32(len(e.Value))
		report.Changed = append(report.Changed, e.Tag.String())
		if e.Tag == (dicom.Tag{Group: 0x0020, Element: 0x000d}) || e.Tag == (dicom.Tag{Group: 0x0020, Element: 0x000e}) || e.Tag == (dicom.Tag{Group: 0x0008, Element: 0x0018}) {
			if len(e.Value) > 0 {
				e.Value = []byte(uid(strings.TrimSpace(string(e.Value))))
				e.Length = uint32(len(e.Value))
			}
		}
		kept = append(kept, e)
	}
	out.Elements = kept
	out.StudyUID, out.SeriesUID, out.SOPInstanceUID = uid(out.StudyUID), uid(out.SeriesUID), uid(out.SOPInstanceUID)
	out.PatientName = "ANON"
	out.PatientID = "ANON-" + hash(in.PatientID)[:12]
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%v", p.ID, p.Version, report.Changed)))
	report.Digest = hex.EncodeToString(sum[:])
	return out, report, nil
}
func hash(v string) string { s := sha256.Sum256([]byte(v)); return hex.EncodeToString(s[:]) }
func shift(v string, days int) string {
	t, err := time.Parse("20060102", strings.TrimSpace(v))
	if err != nil {
		return "19000101"
	}
	return t.AddDate(0, 0, days).Format("20060102")
}

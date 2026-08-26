package repository

import (
	"example.com/dicom-gateway/internal/dicom"
	"sort"
	"strings"
)

type InstanceQuery struct {
	StudyUID  string
	SeriesUID string
	Modality  string
	PatientID string
	Limit     int
	Offset    int
}

func (r *Repository) QueryInstances(q InstanceQuery) ([]dicom.Instance, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := []dicom.Instance{}
	for _, v := range r.instances {
		if q.StudyUID != "" && v.StudyUID != q.StudyUID {
			continue
		}
		if q.SeriesUID != "" && v.SeriesUID != q.SeriesUID {
			continue
		}
		if q.Modality != "" && !strings.EqualFold(v.Metadata.Modality, q.Modality) {
			continue
		}
		if q.PatientID != "" && v.Metadata.PatientID != q.PatientID {
			continue
		}
		all = append(all, v)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.Before(all[j].CreatedAt) })
	total := len(all)
	start := q.Offset
	if start > total {
		start = total
	}
	end := start + q.Limit
	if q.Limit <= 0 {
		end = total
	}
	if end > total {
		end = total
	}
	return all[start:end], total
}

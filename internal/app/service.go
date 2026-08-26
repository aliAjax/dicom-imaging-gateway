package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"example.com/dicom-gateway/internal/audit"
	"example.com/dicom-gateway/internal/deid"
	"example.com/dicom-gateway/internal/dicom"
	"example.com/dicom-gateway/internal/jobs"
	"example.com/dicom-gateway/internal/repository"
	"example.com/dicom-gateway/internal/routing"
	"example.com/dicom-gateway/internal/storage"
	"example.com/dicom-gateway/internal/uid"
	"fmt"
	"io"
	"time"
)

type Service struct {
	Parser    dicom.Parser
	Repo      *repository.Repository
	Store     storage.ObjectStore
	Audit     *audit.Log
	UID       *uid.Mapper
	Routes    *routing.Engine
	Scheduler *jobs.Scheduler
}

func New(p dicom.Parser, repo *repository.Repository, store storage.ObjectStore, a *audit.Log, mapper *uid.Mapper, r *routing.Engine, s *jobs.Scheduler) *Service {
	return &Service{Parser: p, Repo: repo, Store: store, Audit: a, UID: mapper, Routes: r, Scheduler: s}
}
func (s *Service) Validate(ctx context.Context, r io.Reader) (dicom.Dataset, error) {
	return s.Parser.Parse(r)
}

func (s *Service) ValidateWithCodec(ctx context.Context, r io.Reader, codec dicom.Codec) (dicom.Dataset, error) {
	if dicom.IsNilCodec(codec) {
		return dicom.Dataset{}, fmt.Errorf("decode validation input: codec is nil")
	}
	decoded, err := codec.Decode(ctx, mustRead(r))
	if err != nil {
		return dicom.Dataset{}, fmt.Errorf("decode validation input: %w", err)
	}
	return s.Parser.Parse(&bytesReader{b: decoded})
}

func mustRead(r io.Reader) []byte { data, _ := io.ReadAll(r); return data }
func (s *Service) Ingest(ctx context.Context, r io.Reader) (dicom.Instance, error) {
	d, err := s.Parser.Parse(r)
	if err != nil {
		return dicom.Instance{}, err
	}
	if d.SOPInstanceUID == "" {
		return dicom.Instance{}, fmt.Errorf("missing SOP Instance UID")
	}
	key := storage.Key(d.SOPInstanceUID)
	_, err = s.Store.Put(ctx, key, readerDataset(d))
	if err != nil {
		return dicom.Instance{}, err
	}
	inst := dicom.Instance{UID: d.SOPInstanceUID, StudyUID: d.StudyUID, SeriesUID: d.SeriesUID, Metadata: d, ObjectKey: key, Status: "received", CreatedAt: time.Now().UTC(), Version: 1}
	if old, ok := s.Repo.GetInstance(inst.UID); ok {
		return old, nil
	}
	if err = s.Repo.PutInstance(inst); err != nil {
		return dicom.Instance{}, err
	}
	s.Audit.Append("instance_ingested", inst.UID, "system", map[string]string{"sha256": d.SHA256})
	return inst, nil
}
func readerDataset(d dicom.Dataset) io.Reader {
	b, _ := dicom.EncodePart10(d)
	return &bytesReader{b: b}
}

type bytesReader struct{ b []byte }

func (r *bytesReader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}
func (s *Service) Deidentify(inst dicom.Instance, policy deid.Policy) (dicom.Instance, deid.Report, error) {
	d, rep, err := policy.Apply(inst.Metadata, s.UID.Map)
	if err != nil {
		return inst, rep, err
	}
	inst.Metadata = d
	inst.UID = d.SOPInstanceUID
	inst.Status = "deidentified"
	inst.Version++
	if err = s.Repo.PutInstance(inst); err != nil {
		return inst, rep, err
	}
	s.Audit.Append("instance_deidentified", inst.UID, "system", map[string]string{"policy": policy.ID})
	return inst, rep, nil
}
func (s *Service) Route(inst dicom.Instance) {
	for _, r := range s.Routes.Match(inst.Metadata) {
		j := dicom.RouteJob{ID: newID(), InstanceUID: inst.UID, DestinationID: r.DestinationID, Status: "queued", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		s.Scheduler.Enqueue(j)
		s.Audit.Append("route_queued", j.ID, "system", map[string]string{"destination": r.DestinationID})
	}
}
func newID() string { b := make([]byte, 12); _, _ = rand.Read(b); return hex.EncodeToString(b) }

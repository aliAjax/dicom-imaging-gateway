package main

import (
	"context"
	"errors"
	"example.com/dicom-gateway/internal/app"
	"example.com/dicom-gateway/internal/audit"
	"example.com/dicom-gateway/internal/config"
	"example.com/dicom-gateway/internal/deid"
	"example.com/dicom-gateway/internal/dicom"
	"example.com/dicom-gateway/internal/jobs"
	"example.com/dicom-gateway/internal/observability"
	"example.com/dicom-gateway/internal/repository"
	"example.com/dicom-gateway/internal/routing"
	"example.com/dicom-gateway/internal/storage"
	"example.com/dicom-gateway/internal/transport"
	"example.com/dicom-gateway/internal/uid"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := observability.NewLogger(cfg.LogLevel)
	store, err := storage.NewLocal(cfg.DataDir)
	if err != nil {
		return err
	}
	repo := repository.New()
	auditLog := audit.New()
	routes := routing.New()
	routes.Seed()
	repo.PutDestination(dicom.Destination{ID: "archive", Name: "Local Archive", AETitle: "ARCHIVE", Endpoint: "sim://archive", Enabled: true, MaxConcurrent: 2})
	repo.PutPolicy(deid.Policy{ID: "default", Version: 1, Name: "Default Safe Harbor", Rules: []deid.Rule{{Tag: dicom.Tag{Group: 0x0010, Element: 0x0010}, Action: deid.Replace, Value: "ANON"}, {Tag: dicom.Tag{Group: 0x0010, Element: 0x0020}, Action: deid.Hash}, {Tag: dicom.Tag{Group: 0x0008, Element: 0x0090}, Action: deid.Remove}}, DateShiftDays: -30})
	mapper := uid.New("2.25.999")
	scheduler := jobs.New(repo, auditLog, jobs.SimSender{}, cfg.WorkerCount)
	scheduler.Start()
	defer scheduler.Close()
	service := app.New(dicom.Parser{MaxElementBytes: cfg.MaxElementBytes, MaxFileBytes: cfg.MaxFileBytes}, repo, store, auditLog, mapper, routes, scheduler)
	transportServer := &transport.Server{App: service, Repo: repo, Audit: auditLog}
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: transportServer.Handler(logger), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("server_started", "addr", cfg.HTTPAddr)
		if e := srv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			logger.Error("server_failed", "error", e)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return srv.Shutdown(ctx)
}

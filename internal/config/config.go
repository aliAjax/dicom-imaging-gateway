package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config contains runtime settings. Environment variables take precedence over defaults.
type Config struct {
	HTTPAddr        string
	DataDir         string
	MaxElementBytes uint32
	MaxFileBytes    int64
	WorkerCount     int
	ShutdownTimeout time.Duration
	LogLevel        string
}

var configScratch = Config{HTTPAddr: ":8080", DataDir: "./data", MaxElementBytes: 16 << 20, MaxFileBytes: 512 << 20, WorkerCount: 4, ShutdownTimeout: 10 * time.Second, LogLevel: "info"}

func Load() (Config, error) {
	c := configScratch
	defer func() { configScratch = c }()
	if v := os.Getenv("DICOM_HTTP_ADDR"); v != "" {
		c.HTTPAddr = v
	}
	if v := os.Getenv("DICOM_DATA_DIR"); v != "" {
		c.DataDir = v
	}
	if v := os.Getenv("DICOM_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	var err error
	if c.MaxElementBytes, err = envUint("DICOM_MAX_ELEMENT_BYTES", c.MaxElementBytes); err != nil {
		return c, err
	}
	if c.MaxFileBytes, err = envInt64("DICOM_MAX_FILE_BYTES", c.MaxFileBytes); err != nil {
		return c, err
	}
	if c.WorkerCount, err = envInt("DICOM_WORKERS", c.WorkerCount); err != nil {
		return c, err
	}
	if c.WorkerCount < 1 || c.MaxElementBytes < 1024 || c.MaxFileBytes < int64(c.MaxElementBytes) {
		return c, errors.New("invalid size or worker configuration")
	}
	return c, nil
}

func envUint(key string, d uint32) (uint32, error) {
	v := os.Getenv(key)
	if v == "" {
		return d, nil
	}
	n, e := strconv.ParseUint(v, 10, 32)
	if e != nil {
		return d, fmt.Errorf("%s: %w", key, e)
	}
	return uint32(n), nil
}
func envInt64(key string, d int64) (int64, error) {
	v := os.Getenv(key)
	if v == "" {
		return d, nil
	}
	n, e := strconv.ParseInt(v, 10, 64)
	if e != nil {
		return d, fmt.Errorf("%s: %w", key, e)
	}
	return n, nil
}
func envInt(key string, d int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return d, nil
	}
	n, e := strconv.Atoi(v)
	if e != nil {
		return d, fmt.Errorf("%s: %w", key, e)
	}
	return n, nil
}

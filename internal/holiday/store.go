package holiday

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// OfficialURL is the Cabinet Office CSV listing national holidays.
const OfficialURL = "https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv"

const (
	// CheckInterval is how long downloaded data is used before checking for
	// updates.
	CheckInterval = 7 * 24 * time.Hour
	// RetryInterval is how long to wait after a failed update check before
	// trying again, while previously downloaded data remains usable.
	RetryInterval = 24 * time.Hour

	dataFile = "syukujitsu.csv"
	metaFile = "syukujitsu.json"

	maxDataSize = 1 << 20
)

// Store keeps a copy of the official holiday CSV in Dir.
type Store struct {
	Dir    string
	URL    string
	Client *http.Client
	Now    func() time.Time
}

type metadata struct {
	CheckedAt    time.Time `json:"checked_at"`
	AttemptedAt  time.Time `json:"attempted_at"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"last_modified,omitempty"`
}

// Load returns the stored holidays, downloading them first when no usable
// local data exists. On failure it returns an empty Set with the error.
func (s *Store) Load(ctx context.Context) (Set, error) {
	if set, err := s.readData(); err == nil {
		return set, nil
	}
	if err := s.update(ctx); err != nil {
		return nil, err
	}
	return s.readData()
}

// Refresh checks the official source for updates when the local data is
// due for a check.
func (s *Store) Refresh(ctx context.Context) error {
	if !s.needsUpdate() {
		return nil
	}
	return s.update(ctx)
}

func (s *Store) needsUpdate() bool {
	if _, err := s.readData(); err != nil {
		return true
	}
	meta := s.readMeta()
	now := s.Now()
	return now.Sub(meta.CheckedAt) >= CheckInterval && now.Sub(meta.AttemptedAt) >= RetryInterval
}

func (s *Store) update(ctx context.Context) error {
	meta := s.readMeta()
	meta.AttemptedAt = s.Now()
	err := s.fetch(ctx, &meta)
	if err == nil {
		meta.CheckedAt = meta.AttemptedAt
	}
	if werr := s.writeMeta(meta); err == nil {
		err = werr
	}
	return err
}

func (s *Store) fetch(ctx context.Context, meta *metadata) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return err
	}
	if _, err := s.readData(); err == nil {
		if meta.ETag != "" {
			req.Header.Set("If-None-Match", meta.ETag)
		}
		if meta.LastModified != "" {
			req.Header.Set("If-Modified-Since", meta.LastModified)
		}
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return nil
	case http.StatusOK:
	default:
		return fmt.Errorf("download holiday data: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDataSize+1))
	if err != nil {
		return err
	}
	if len(body) > maxDataSize {
		return errors.New("holiday data is too large")
	}
	if _, err := Parse(bytes.NewReader(body)); err != nil {
		return err
	}
	if err := s.writeFile(dataFile, body); err != nil {
		return err
	}
	meta.ETag = resp.Header.Get("ETag")
	meta.LastModified = resp.Header.Get("Last-Modified")
	return nil
}

func (s *Store) readData() (Set, error) {
	f, err := os.Open(filepath.Join(s.Dir, dataFile))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

func (s *Store) readMeta() metadata {
	var meta metadata
	if b, err := os.ReadFile(filepath.Join(s.Dir, metaFile)); err == nil {
		// Corrupt metadata only causes an early update check.
		_ = json.Unmarshal(b, &meta)
	}
	return meta
}

func (s *Store) writeMeta(meta metadata) error {
	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return s.writeFile(metaFile, b)
}

// writeFile replaces the named file atomically so that concurrent
// invocations never read partial data.
func (s *Store) writeFile(name string, data []byte) error {
	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.Dir, name+".*.tmp")
	if err != nil {
		return err
	}
	// Removing the temporary file fails harmlessly once it has been renamed.
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(s.Dir, name))
}

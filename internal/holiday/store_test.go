package holiday

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

type fakeServer struct {
	*httptest.Server
	requests atomic.Int32
	status   atomic.Int32
	lastINM  atomic.Value
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	fs := &fakeServer{}
	fs.status.Store(http.StatusOK)
	fs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.requests.Add(1)
		fs.lastINM.Store(r.Header.Get("If-None-Match"))
		status := int(fs.status.Load())
		if status == http.StatusOK && r.Header.Get("If-None-Match") == `"v1"` {
			status = http.StatusNotModified
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		_, _ = w.Write([]byte(sampleCSV))
	}))
	t.Cleanup(fs.Close)
	return fs
}

func newTestStore(t *testing.T, url string, now *time.Time) *Store {
	t.Helper()
	return &Store{
		Dir:    filepath.Join(t.TempDir(), "jcal"),
		URL:    url,
		Client: http.DefaultClient,
		Now:    func() time.Time { return *now },
	}
}

func TestStoreDownloadsWhenNoDataExists(t *testing.T) {
	srv := newFakeServer(t)
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	s := newTestStore(t, srv.URL, &now)

	set, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !set.IsHoliday(2026, time.November, 3) {
		t.Error("expected downloaded holidays")
	}
	if _, err := os.Stat(filepath.Join(s.Dir, dataFile)); err != nil {
		t.Errorf("data not stored: %v", err)
	}
}

func TestStoreReusesCachedData(t *testing.T) {
	srv := newFakeServer(t)
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	s := newTestStore(t, srv.URL, &now)

	if _, err := s.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	now = now.Add(24 * time.Hour)
	set, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !set.IsHoliday(2026, time.November, 3) {
		t.Error("expected cached holidays")
	}
	if got := srv.requests.Load(); got != 1 {
		t.Errorf("requests = %d, want 1", got)
	}
}

func TestStoreRefreshesStaleDataConditionally(t *testing.T) {
	srv := newFakeServer(t)
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	s := newTestStore(t, srv.URL, &now)
	if _, err := s.Load(context.Background()); err != nil {
		t.Fatal(err)
	}

	now = now.Add(CheckInterval)
	if err := s.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := srv.requests.Load(); got != 2 {
		t.Fatalf("requests = %d, want 2", got)
	}
	if got := srv.lastINM.Load(); got != `"v1"` {
		t.Errorf("If-None-Match = %v", got)
	}

	// A 304 response counts as a successful check.
	now = now.Add(time.Hour)
	if err := s.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := srv.requests.Load(); got != 2 {
		t.Errorf("requests = %d, want 2", got)
	}
}

func TestStoreKeepsDataWhenUpdateFails(t *testing.T) {
	srv := newFakeServer(t)
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	s := newTestStore(t, srv.URL, &now)
	if _, err := s.Load(context.Background()); err != nil {
		t.Fatal(err)
	}

	srv.status.Store(http.StatusInternalServerError)
	now = now.Add(CheckInterval)
	if err := s.Refresh(context.Background()); err == nil {
		t.Error("expected refresh error")
	}
	set, err := s.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !set.IsHoliday(2026, time.November, 3) {
		t.Error("previous data should be kept")
	}

	// Failed checks are retried later, but not on every invocation.
	if err := s.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := srv.requests.Load(); got != 2 {
		t.Errorf("requests = %d, want 2", got)
	}
	now = now.Add(RetryInterval)
	srv.status.Store(http.StatusOK)
	if err := s.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := srv.requests.Load(); got != 3 {
		t.Errorf("requests = %d, want 3", got)
	}
}

func TestStoreWithoutNetworkOrData(t *testing.T) {
	srv := newFakeServer(t)
	srv.Close()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	s := newTestStore(t, srv.URL, &now)

	set, err := s.Load(context.Background())
	if err == nil {
		t.Error("expected error")
	}
	if set.IsHoliday(2026, time.November, 3) {
		t.Error("expected no holidays")
	}
	if _, err := os.Stat(filepath.Join(s.Dir, dataFile)); !os.IsNotExist(err) {
		t.Errorf("no data should be stored: %v", err)
	}
}

func TestStoreRejectsInvalidDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html>maintenance</html>"))
	}))
	defer srv.Close()
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	s := newTestStore(t, srv.URL, &now)

	if _, err := s.Load(context.Background()); err == nil {
		t.Error("expected error")
	}
	if _, err := os.Stat(filepath.Join(s.Dir, dataFile)); !os.IsNotExist(err) {
		t.Errorf("invalid data should not be stored: %v", err)
	}
}

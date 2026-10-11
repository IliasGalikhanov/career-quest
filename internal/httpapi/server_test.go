package httpapi

import (
	"bytes"
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"careerquest/internal/store"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
)

func request(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func decodeDataset(t *testing.T, w *httptest.ResponseRecorder) store.Dataset {
	t.Helper()
	if w.Code >= 300 {
		t.Fatalf("HTTP %d: %s", w.Code, w.Body.String())
	}
	var d store.Dataset
	if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestWorkflowSurvivesRestartAndRejectsStaleWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	h := New(db, "")
	d := decodeDataset(t, request(t, h, "POST", "/api/demo", nil))
	prefix := "/api/datasets/" + d.ID
	d = decodeDataset(t, request(t, h, "POST", prefix+"/participation", map[string]any{"revision": d.Revision, "employee_id": "person-01", "event_id": "architecture-lab"}))
	record := d.Bundle.History[len(d.Bundle.History)-1].ID
	stale := request(t, h, "PUT", prefix+"/participation/"+record, map[string]any{"revision": 1, "status": "completed"})
	if stale.Code != 409 {
		t.Fatalf("expected conflict, got %d", stale.Code)
	}
	d = decodeDataset(t, request(t, h, "PUT", prefix+"/participation/"+record, map[string]any{"revision": d.Revision, "status": "completed"}))
	d = decodeDataset(t, request(t, h, "PUT", prefix+"/participation/"+record, map[string]any{"revision": d.Revision, "status": "completed"}))
	a, _ := domain.Analyze(d.Bundle, "person-01")
	if a.Levels["architecture"] != 3 {
		t.Fatalf("completion duplicated: %v", a.Levels)
	}
	if w := request(t, h, "GET", prefix+"/employees/person-01/analysis", nil); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h = New(db, "")
	reopened := decodeDataset(t, request(t, h, "GET", prefix, nil))
	if reopened.Revision != d.Revision {
		t.Fatal("revision not persisted")
	}
	a, _ = domain.Analyze(reopened.Bundle, "person-01")
	if a.Levels["architecture"] != 3 {
		t.Fatal("skill progress lost on restart")
	}
	w := request(t, h, "GET", prefix+"/evidence", nil)
	var evidence []any
	if err := json.Unmarshal(w.Body.Bytes(), &evidence); err != nil || len(evidence) != 1 {
		t.Fatal("recommendation evidence not persisted")
	}
}

func TestImportPreviewIsReadOnlyAndInvalidCommitIsAtomic(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := New(db, "")
	upload := func(path string, b domain.Bundle) *httptest.ResponseRecorder {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("files", "demo.json")
		if err != nil {
			t.Fatal(err)
		}
		if err := json.NewEncoder(part).Encode(b); err != nil {
			t.Fatal(err)
		}
		_ = writer.WriteField("synthetic", "true")
		writer.Close()
		r := httptest.NewRequest("POST", path, &body)
		r.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := upload("/api/import/preview", demo.Dataset()); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	rows, _ := db.List(t.Context())
	if len(rows) != 0 {
		t.Fatal("preview wrote data")
	}
	bad := demo.Dataset()
	bad.Employees[0].Skills["unknown"] = 9
	if w := upload("/api/import/commit", bad); w.Code != 422 {
		t.Fatal(w.Body.String())
	}
	rows, _ = db.List(t.Context())
	if len(rows) != 0 {
		t.Fatal("invalid import wrote data")
	}
	if w := upload("/api/import/commit", demo.Dataset()); w.Code != 201 {
		t.Fatal(w.Body.String())
	}
}

func TestOnlyOneConcurrentStartWins(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := New(db, "")
	d := decodeDataset(t, request(t, h, "POST", "/api/demo", nil))
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := request(t, h, "POST", "/api/datasets/"+d.ID+"/participation", map[string]any{"revision": d.Revision, "employee_id": "person-01", "event_id": "architecture-lab"})
			results <- w.Code
		}()
	}
	wg.Wait()
	close(results)
	codes := map[int]int{}
	for code := range results {
		codes[code]++
	}
	if codes[200] != 1 || codes[409] != 1 {
		t.Fatalf("unexpected results: %v", codes)
	}
}

func TestCrossOriginWriteRejected(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := httptest.NewRequest("POST", "http://localhost/api/demo", nil)
	r.Header.Set("Origin", "https://untrusted.example")
	w := httptest.NewRecorder()
	New(db, "").ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross-origin write allowed")
	}
}

func TestManageDatasetsAndRequestDiagnostics(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "manage.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := New(db, "")
	d := decodeDataset(t, request(t, h, "POST", "/api/demo", nil))
	prefix := "/api/datasets/" + d.ID
	d = decodeDataset(t, request(t, h, "PATCH", prefix, map[string]any{"revision": d.Revision, "name": "Renamed workspace"}))
	if d.Name != "Renamed workspace" || d.Bundle.Meta.Dataset != d.Name {
		t.Fatal("name mismatch")
	}
	stale := request(t, h, "DELETE", prefix, map[string]any{"revision": 1})
	if stale.Code != 409 {
		t.Fatal(stale.Code)
	}
	var problem apiError
	if err = json.Unmarshal(stale.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem.RequestID == "" || problem.RequestID != stale.Header().Get("X-Request-ID") {
		t.Fatal("missing correlated request ID")
	}
	if w := request(t, h, "DELETE", prefix, map[string]any{"revision": d.Revision}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := request(t, h, "GET", prefix, nil); w.Code != 404 {
		t.Fatal("deleted dataset remains")
	}
}

func TestReflectionSurvivesExportImport(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "reflection.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := New(db, "")
	d := decodeDataset(t, request(t, h, "POST", "/api/demo", nil))
	prefix := "/api/datasets/" + d.ID
	d = decodeDataset(t, request(t, h, "POST", prefix+"/participation", map[string]any{"revision": d.Revision, "employee_id": "person-01", "event_id": "architecture-lab"}))
	record := d.Bundle.History[len(d.Bundle.History)-1].ID
	d = decodeDataset(t, request(t, h, "PUT", prefix+"/participation/"+record, map[string]any{"revision": d.Revision, "status": "completed", "reflection": "Compared retries and idempotency in a component diagram."}))
	exported := request(t, h, "GET", prefix+"/export", nil)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, _ := writer.CreateFormFile("files", "backup.json")
	file.Write(exported.Body.Bytes())
	writer.WriteField("synthetic", "true")
	writer.Close()
	req := httptest.NewRequest("POST", "/api/import/commit", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	out := httptest.NewRecorder()
	h.ServeHTTP(out, req)
	restored := decodeDataset(t, out)
	if restored.Bundle.History[len(restored.Bundle.History)-1].Reflection != d.Bundle.History[len(d.Bundle.History)-1].Reflection {
		t.Fatal("reflection lost")
	}
	actual, _ := json.Marshal(restored.Bundle)
	expected, _ := json.Marshal(d.Bundle)
	if !bytes.Equal(actual, expected) {
		t.Fatal("backup round trip changed domain state")
	}
}

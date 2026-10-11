package httpapi

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"careerquest/internal/importer"
	"careerquest/internal/store"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	store  *store.Store
	webDir string
}
type apiError struct {
	Code      string         `json:"code"`
	RequestID string         `json:"request_id"`
	Issues    []domain.Issue `json:"issues,omitempty"`
}

func New(db *store.Store, webDir string) http.Handler {
	s := &Server{db, webDir}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/datasets", s.list)
	mux.HandleFunc("POST /api/demo", s.demo)
	mux.HandleFunc("GET /api/example", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="career-quest-demo.json"`)
		respond(w, 200, demo.Dataset())
	})
	mux.HandleFunc("POST /api/import/preview", s.importData)
	mux.HandleFunc("POST /api/import/commit", s.importData)
	mux.HandleFunc("GET /api/datasets/{id}", s.dataset)
	mux.HandleFunc("PATCH /api/datasets/{id}", s.rename)
	mux.HandleFunc("DELETE /api/datasets/{id}", s.delete)
	mux.HandleFunc("GET /api/datasets/{id}/export", s.export)
	mux.HandleFunc("GET /api/datasets/{id}/evidence", s.evidence)
	mux.HandleFunc("GET /api/datasets/{id}/employees/{employee}/analysis", s.analysis)
	mux.HandleFunc("PUT /api/datasets/{id}/employees/{employee}/goal", s.goal)
	mux.HandleFunc("PUT /api/datasets/{id}/settings", s.settings)
	mux.HandleFunc("POST /api/datasets/{id}/participation", s.start)
	mux.HandleFunc("PUT /api/datasets/{id}/participation/{record}", s.finish)
	mux.HandleFunc("GET /api/datasets/{id}/hr", s.hr)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) { fail(w, 404, "not_found", nil) })
	mux.HandleFunc("/", s.static)
	return diagnostics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		if r.Method != "GET" && r.Method != "HEAD" {
			origin := r.Header.Get("Origin")
			if origin != "" {
				u, err := url.Parse(origin)
				if err != nil || u.Host != r.Host {
					fail(w, 403, "origin_rejected", nil)
					return
				}
			}
			if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				fail(w, 403, "origin_rejected", nil)
				return
			}
		}
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("request panic", "path", r.URL.Path)
				fail(w, 500, "internal", nil)
			}
		}()
		mux.ServeHTTP(w, r)
	}))
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("write response", "error", err)
	}
}
func fail(w http.ResponseWriter, status int, code string, issues []domain.Issue) {
	respond(w, status, apiError{Code: code, Issues: issues, RequestID: w.Header().Get("X-Request-ID")})
}
func readJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		fail(w, 400, "invalid_json", nil)
		return false
	}
	if err := d.Decode(new(any)); err != io.EOF {
		fail(w, 400, "invalid_json", nil)
		return false
	}
	return true
}
func handleError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, store.ErrNotFound), errors.Is(err, domain.ErrActivityNotFound):
		fail(w, 404, "not_found", nil)
	case errors.Is(err, store.ErrConflict):
		fail(w, 409, "revision_conflict", nil)
	case errors.Is(err, domain.ErrIneligible):
		fail(w, 422, "ineligible", nil)
	case errors.Is(err, domain.ErrFutureSession):
		fail(w, 422, "future_session", nil)
	case errors.Is(err, domain.ErrInvalidTransition):
		fail(w, 422, "invalid_transition", nil)
	default:
		slog.Error("request failed", "error", err)
		fail(w, 500, "internal", nil)
	}
	return true
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if handleError(w, s.store.Ping(r.Context())) {
		return
	}
	respond(w, 200, map[string]string{"status": "ok", "algorithm": domain.AlgorithmVersion})
}
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	out, err := s.store.List(r.Context())
	if !handleError(w, err) {
		respond(w, 200, out)
	}
}
func (s *Server) demo(w http.ResponseWriter, r *http.Request) {
	d, err := s.store.Create(r.Context(), demo.Dataset())
	if !handleError(w, err) {
		respond(w, 201, d)
	}
}
func (s *Server) dataset(w http.ResponseWriter, r *http.Request) {
	d, err := s.store.Get(r.Context(), r.PathValue("id"))
	if !handleError(w, err) {
		respond(w, 200, d)
	}
}
func (s *Server) export(w http.ResponseWriter, r *http.Request) {
	d, err := s.store.Get(r.Context(), r.PathValue("id"))
	if handleError(w, err) {
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="career-quest-backup.json"`)
	respond(w, 200, d.Bundle)
}
func (s *Server) evidence(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.Runs(r.Context(), r.PathValue("id"))
	if handleError(w, err) {
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="recommendation-evidence.json"`)
	respond(w, 200, rows)
}
func (s *Server) analysis(w http.ResponseWriter, r *http.Request) {
	d, err := s.store.Get(r.Context(), r.PathValue("id"))
	if handleError(w, err) {
		return
	}
	a, ok := domain.Analyze(d.Bundle, r.PathValue("employee"))
	if !ok {
		fail(w, 404, "not_found", nil)
		return
	}
	if handleError(w, s.store.SaveRun(r.Context(), d, a)) {
		return
	}
	respond(w, 200, a)
}
func (s *Server) goal(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Revision int          `json:"revision"`
		Goal     *domain.Goal `json:"goal"`
	}
	if !readJSON(w, r, &request) {
		return
	}
	d, err := s.store.Update(r.Context(), r.PathValue("id"), request.Revision, func(b *domain.Bundle) error {
		if request.Goal != nil {
			found := false
			for _, p := range b.RoleProfiles {
				if p.Role == request.Goal.Role && p.Grade == request.Goal.Grade {
					found = true
				}
			}
			if !found {
				return domain.ErrIneligible
			}
		}
		for i, e := range b.Employees {
			if e.ID == r.PathValue("employee") {
				b.Employees[i].Goal = request.Goal
				return nil
			}
		}
		return domain.ErrActivityNotFound
	})
	if !handleError(w, err) {
		respond(w, 200, d)
	}
}
func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Revision int             `json:"revision"`
		Settings domain.Settings `json:"settings"`
	}
	if !readJSON(w, r, &request) {
		return
	}
	if issues := domain.ValidateSettings(request.Settings); len(issues) > 0 {
		fail(w, 422, "validation", issues)
		return
	}
	d, err := s.store.Update(r.Context(), r.PathValue("id"), request.Revision, func(b *domain.Bundle) error {
		if request.Settings.Date < b.Meta.AsOfDate {
			return domain.ErrIneligible
		}
		for _, h := range b.History {
			if h.CompletedAt > request.Settings.Date {
				return domain.ErrFutureSession
			}
		}
		b.Settings = request.Settings
		return nil
	})
	if !handleError(w, err) {
		respond(w, 200, d)
	}
}
func (s *Server) start(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Revision   int    `json:"revision"`
		EmployeeID string `json:"employee_id"`
		EventID    string `json:"event_id"`
	}
	if !readJSON(w, r, &request) {
		return
	}
	d, err := s.store.Update(r.Context(), r.PathValue("id"), request.Revision, func(b *domain.Bundle) error {
		return domain.Start(b, request.EmployeeID, request.EventID, store.NewID())
	})
	if !handleError(w, err) {
		respond(w, 200, d)
	}
}
func (s *Server) finish(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Revision   int    `json:"revision"`
		Status     string `json:"status"`
		Reflection string `json:"reflection"`
	}
	if !readJSON(w, r, &request) {
		return
	}
	d, err := s.store.Update(r.Context(), r.PathValue("id"), request.Revision, func(b *domain.Bundle) error {
		return domain.FinishWithReflection(b, r.PathValue("record"), request.Status, request.Reflection)
	})
	if !handleError(w, err) {
		respond(w, 200, d)
	}
}
func (s *Server) importData(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		fail(w, 400, "invalid_upload", nil)
		return
	}
	defer r.MultipartForm.RemoveAll()
	if r.FormValue("synthetic") != "true" {
		fail(w, 422, "synthetic_required", nil)
		return
	}
	files := map[string][]byte{}
	for _, headers := range r.MultipartForm.File {
		for _, h := range headers {
			if _, exists := files[h.Filename]; exists {
				fail(w, 422, "duplicate_file", nil)
				return
			}
			f, err := h.Open()
			if err != nil {
				fail(w, 400, "invalid_upload", nil)
				return
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				fail(w, 400, "invalid_upload", nil)
				return
			}
			files[h.Filename] = data
		}
	}
	if len(files) == 0 {
		fail(w, 400, "invalid_upload", nil)
		return
	}
	var base *domain.Bundle
	var existing store.Dataset
	if r.FormValue("append_to") != "" {
		var err error
		existing, err = s.store.Get(r.Context(), r.FormValue("append_to"))
		if handleError(w, err) {
			return
		}
		base = &existing.Bundle
	}
	b, issues := importer.Parse(files, base)
	if len(issues) > 0 {
		fail(w, 422, "validation", issues)
		return
	}
	if strings.HasSuffix(r.URL.Path, "preview") {
		respond(w, 200, map[string]any{"name": b.Meta.Dataset, "employees": len(b.Employees), "events": len(b.Events), "skills": len(b.Skills), "history": len(b.History), "as_of_date": b.Meta.AsOfDate, "uncertain_dates": uncertainCount(b)})
		return
	}
	var d store.Dataset
	var err error
	if base == nil {
		d, err = s.store.Create(r.Context(), b)
	} else {
		revision, _ := strconv.Atoi(r.FormValue("revision"))
		d, err = s.store.Update(r.Context(), existing.ID, revision, func(current *domain.Bundle) error { *current = b; return nil })
	}
	if !handleError(w, err) {
		respond(w, 201, d)
	}
}
func uncertainCount(b domain.Bundle) int {
	n := 0
	for _, e := range b.Employees {
		_, count := domain.EffectiveLevels(b, e)
		n += count
	}
	return n
}
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		w.WriteHeader(405)
		return
	}
	path := filepath.Join(s.webDir, filepath.FromSlash(strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), "/")))
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
	if strings.Contains(filepath.Base(r.URL.Path), ".") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, filepath.Join(s.webDir, "index.html"))
}

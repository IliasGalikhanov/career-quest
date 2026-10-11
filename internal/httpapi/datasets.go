package httpapi

import (
	"careerquest/internal/domain"
	"net/http"
	"strings"
	"unicode/utf8"
)

func (s *Server) rename(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Revision int    `json:"revision"`
		Name     string `json:"name"`
	}
	if !readJSON(w, r, &request) {
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" || utf8.RuneCountInString(request.Name) > 120 {
		fail(w, 422, "validation", []domain.Issue{{Path: "name", Code: "range", Detail: "Expected 1–120 characters"}})
		return
	}
	d, err := s.store.Update(r.Context(), r.PathValue("id"), request.Revision, func(b *domain.Bundle) error { b.Meta.Dataset = request.Name; return nil })
	if !handleError(w, err) {
		respond(w, 200, d)
	}
}
func (s *Server) delete(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Revision int `json:"revision"`
	}
	if !readJSON(w, r, &request) {
		return
	}
	if !handleError(w, s.store.Delete(r.Context(), r.PathValue("id"), request.Revision)) {
		respond(w, 200, map[string]bool{"deleted": true})
	}
}

package handler

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type WebHandler struct {
	templates *template.Template
}

func NewWebHandler(templateDir string) (*WebHandler, error) {
	templates, err := template.ParseFiles(filepath.Join(templateDir, "index.html"))
	if err != nil {
		return nil, err
	}

	return &WebHandler{
		templates: templates,
	}, nil
}

func (h *WebHandler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

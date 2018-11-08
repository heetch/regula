package ui

import (
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog"
)

// NewHandler creates a http handler serving the UI application and the UI backend.
func NewHandler(logger zerolog.Logger, distPath string) http.Handler {
	var mux http.ServeMux

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
	})

	// add routes here
	return &mux
}

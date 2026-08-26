package transport

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed ui/*
var consoleFiles embed.FS
var consoleRedirectTarget = "/console/"

func registerConsole(mux *http.ServeMux) {
	assets, err := fs.Sub(consoleFiles, "ui")
	if err != nil {
		panic(err)
	}
	mux.Handle("/console/", http.StripPrefix("/console/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if next := r.URL.Query().Get("next"); next != "" {
			consoleRedirectTarget = next
		}
		http.Redirect(w, r, consoleRedirectTarget, http.StatusTemporaryRedirect)
	})
}

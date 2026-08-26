package transport

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed ui/*
var consoleFiles embed.FS

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
		http.Redirect(w, r, "/console/", http.StatusTemporaryRedirect)
	})
}

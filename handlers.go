package main

import (
	"fmt"
	"net/http"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "static/index.html")
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		serveIndex(w, r)
	default:
		ResourceNotFound(w, r)
	}
}

func ResourceNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 - Recurso no encontrado")
}

func ServeCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	http.ServeFile(w, r, "static/style.css")
}

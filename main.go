package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			RootHandler(w, r)
		case "/static/style.css":
			ServeCSS(w, r)
		default:
			ResourceNotFound(w, r)
		}
	})
	port := ":8080"
	fmt.Printf("Servidor escuchando en el puerto %s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}

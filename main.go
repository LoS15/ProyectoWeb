package main

import (
	db "ProyectoWeb/db/sqlc"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

// Variables globales para poder acceder desde handlers.go
var (
	dbConn  *sql.DB
	queries *db.Queries
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

	// Conexión a la base de datos
	connStr := "host=localhost port=5432 user=user password=user_password dbname=proyecto_web sslmode=disable"
	var err error
	dbConn, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer dbConn.Close()

	// Instancio el repositorio
	queries = db.New(dbConn)

	// Contexto necesario para las operaciones de sqlc
	ctx := context.Background()

	port := ":8080"
	fmt.Printf("Servidor escuchando en el puerto %s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}

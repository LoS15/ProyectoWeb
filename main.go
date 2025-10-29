package main

import (
	db "ProyectoWeb/db/sqlc"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

// Variables globales para poder acceder desde handlers.go
var (
	dbConn  *sql.DB
	queries *db.Queries
)

func main() {
	http.HandleFunc("/", ServeIndex)
	http.HandleFunc("/partidos", PartidosHandler)
	http.HandleFunc("/partidos/", PartidosHandler)
	http.HandleFunc("/usuarios/", UsuariosHandler)
	http.HandleFunc("/usuarios", UsuariosHandler)
	http.HandleFunc("/estadisticas-jugador", EstadisticasJugadorHandler)
	http.HandleFunc("/estadisticas-arquero", EstadisticasArqueroHandler)
	http.HandleFunc("/crearPartido", crearPartidoCompleto)
	//ResourceNotFound(w, r)
	//http.HandleFunc("/crearPartido", func)
	//http.HandleFunc("/estadisticas", func)
	/*http.
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			ServeIndex(w, r)
		case strings.HasPrefix(r.URL.Path, "/usuarios"):
			UsuariosHandler(w, r)
		case "/usuarios":
			UsuariosHandler(w, r)
		case "/partidos":
			PartidosHandler(w, r)
		case "/estadisticas-jugador":
			EstadisticasJugadorHandler(w, r)
		case "/estadistica-arquero":
			EstadisticasArqueroHandler(w, r)
		case "/estadisticas":
			//Obtener estadistica que corresponde o modificar estadistica que corresponde desde la web
		case "/crearPartido":
			//PartidoCompletoHandler()
		default:
			ResourceNotFound(w, r)
		}
	})*/

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

	port := ":8080"
	fmt.Printf("Servidor escuchando en el puerto %s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}

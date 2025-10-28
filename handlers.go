package main

import (
	db "ProyectoWeb/db/sqlc"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Tipos auxiliares para las funciones de los handlers de Partido
type CrearPartidoCompletoRequest struct {
	IDUsuario          int32                              `json:"id_usuario"`
	Fecha              time.Time                          `json:"fecha"`
	Cancha             string                             `json:"cancha"`
	Puntuacion         int32                              `json:"puntuacion"`
	TipoEstadistica    string                             `json:"tipo_estadistica"`
	EstadisticaJugador *db.InsertEstadisticaJugadorParams `json:"estadisticas_jugador,omitempty"`
	EstadisticaArquero *db.InsertEstadisticaArqueroParams `json:"estadistica_arquero,omitempty"`
}

// Tipos para poder hacer chequeos de nulos en tipos primitivos que no permiten control de nulos
type PartidoRequest struct {
	IDUsuario  *int32     `json:"id_usuario"` // puntero para saber si vino
	Fecha      *time.Time `json:"fecha"`      // puntero para poder usar IsZero o nil
	Cancha     string     `json:"cancha"`
	Puntuacion *int32     `json:"puntuacion"` // puntero para saber si vino
}
type EstadisticaJugadorRequest struct {
	Goles            *int32 `json:"goles"`       // puntero para saber si vino
	Asistencias      *int32 `json:"asistencias"` // puntero para saber si vino
	PasesCompletados string `json:"pases_completados,omitempty"`
	DuelosGanados    string `json:"duelos_ganados,omitempty"`
}
type EstadisticaArqueroRequest struct {
	GolesRecibidos    int32  `json:"goles_recibidos"` // puntero para saber si vino
	AtajadasClave     int32  `json:"atajadas_clave"`  // puntero para saber si vino
	SaquesCompletados string `json:"saques_completados"`
}

// Funcion para el handler POST partido
func crearPartidoCompletoHandler(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request en una entidad intermedia
	var request CrearPartidoCompletoRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON, primero del partido
	partidoReq := PartidoRequest{
		IDUsuario:  &request.IDUsuario,
		Fecha:      &request.Fecha,
		Cancha:     request.Cancha,
		Puntuacion: &request.Puntuacion,
	}
	err = validarPartido(partidoReq)
	if err != nil {
		// Si hay error en la validación del partido lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Uso transaccion por que son dos operaciones separadas en la BD y deben hacerse o ambas o ninguna
	// Comienzo la transaccion con la base de datos
	transaccion, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		// Si ocurre un error al iniciar la transaccion, lanzo un código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error iniciando la transaccion", http.StatusInternalServerError)
		return
	}
	// Retrasamos el rollback hasta el final de la ejecución, en caso de que haya algún error
	defer transaccion.Rollback()

	// Creo el objeto que me permite ejecutar todas las consultas dentro de la transaccion
	qtran := queries.WithTx(transaccion)

	// Separo la informacion de la entidad intermedia
	// Creo el partido
	datosNuevoPartido := db.InsertPartidoParams{
		IDUsuario:  request.IDUsuario,
		Fecha:      request.Fecha,
		Cancha:     request.Cancha,
		Puntuacion: request.Puntuacion,
	}

	// Inserto en la tabla Partido el nuevo partido
	nuevoPartido, err := qtran.InsertPartido(ctx, datosNuevoPartido)
	if err != nil {
		// Si ocurre un error al insertar el nuevo partido, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error creando partido", http.StatusInternalServerError)
		return
	}

	// Creo las estadisticas del partido, segun el tipo
	switch request.TipoEstadistica {
	case "jugador":
		// Si el tipo de estadisticas es de jugador de campo

		if request.EstadisticaJugador == nil {
			// Si las estadisticas no están cargadas, lanzo código 400 y finalizo la ejecucion del handler
			http.Error(w, "Faltan datos de estadisticas de jugador", http.StatusBadRequest)
			return
		}

		// Creo las estadisticas si es jugador
		datosNuevaEstadisticaJugador := db.InsertEstadisticaJugadorParams{
			IDUsuario:        nuevoPartido.IDUsuario,
			IDPartido:        nuevoPartido.IDPartido,
			Goles:            request.EstadisticaJugador.Goles,
			Asistencias:      request.EstadisticaJugador.Asistencias,
			PasesCompletados: request.EstadisticaJugador.PasesCompletados,
			DuelosGanados:    request.EstadisticaJugador.DuelosGanados,
		}

		_, err := qtran.InsertEstadisticaJugador(ctx, datosNuevaEstadisticaJugador)
		if err != nil {
			// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
			http.Error(w, "Error iniciando la transaccion", http.StatusInternalServerError)
			return
		}

	case "arquero":
		// Si el tipo de estadisticas es de arquero

		if request.EstadisticaArquero == nil {
			// Si las estadisticas no están cargadas, lanzo código 400 y finalizo la ejecucion del handler
			http.Error(w, "Faltan datos de estadisticas de arquero", http.StatusBadRequest)
			return
		}

		// Creo las estadisticas si es arquero
		datosNuevaEstadisticaArquero := db.InsertEstadisticaArqueroParams{
			IDUsuario:         nuevoPartido.IDUsuario,
			IDPartido:         nuevoPartido.IDPartido,
			GolesRecibidos:    request.EstadisticaArquero.GolesRecibidos,
			AtajadasClave:     request.EstadisticaArquero.AtajadasClave,
			SaquesCompletados: request.EstadisticaArquero.SaquesCompletados,
		}

		_, err := qtran.InsertEstadisticaArquero(ctx, datosNuevaEstadisticaArquero)
		if err != nil {
			// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
			http.Error(w, "Error iniciando la transaccion", http.StatusInternalServerError)
			return
		}
	default:
		// Si el tipo de estadisticas es otro, lanzo código 400 y finalizo la ejecucion del handler
		http.Error(w, "Tipo de estadística no válido", http.StatusBadRequest)
		return
	}

	// Confirmo la transaccion
	err = transaccion.Commit()
	if err != nil {
		// Si ocurre un error al confirmar la transaccion, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error confirmando transacción", http.StatusInternalServerError)
		return
	}

	// Establezo el header de respuesta de tipo JSON, lanzo el código 201 y respondo con el partido y sus estadisticas como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(request)
}

// Funcion para el handler GET todos los partidos
func listarTodosLosPartidos(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Obtengo todos los partidos de la tabla Partido
	partidos, err := queries.GetAllPartido(ctx)
	if err != nil {
		// Si ocurre un error obteniendo todos los partidos, lanzo código 404 y finalizo la ejecucion del handler
		http.Error(w, "Error obteniendo todos los partidos existentes", http.StatusNotFound)
		return
	}

	// Establezo el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(partidos)
}

// Funcion para el handler GET todos los partidos para un usuario
func listarTodosLosPartidosPorUsuario(w http.ResponseWriter, r *http.Request, id_usuario int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Obtengo todos los partidos para un usuario dado de la tabla Partido
	partidos, err := queries.ListPartidosPorUsuario(ctx, id_usuario)
	if err != nil {
		// Si ocurre un error obteniendo todos los partidos para un usuario dado, lanzo código 404 y finalizo la ejecucion del handler
		http.Error(w, fmt.Sprintf("Error obteniendo todos los partidos para el usuario %d", id_usuario), http.StatusNotFound)
		return
	}

	// Establezo el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados para el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(partidos)
}

// Funcion para el handler GET partido dado para un usuario dado
func listarPartidoPorUsuario(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Crea el tipo auxiliar necesario para la operación sqlc
	informacionPartido := db.GetPartidoPorUsuarioParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Obtengo un partido dado para un usuario dado
	partido, err := queries.GetPartidoPorUsuario(ctx, informacionPartido)
	if err != nil {
		// Si ocurre un error obteniendo el partido dado para el usuario dado, lanzo código 404 y finalizo la ejecucion del handler
		http.Error(w, fmt.Sprintf("Error obteniendo el partido %d para el usuario %d", id_partido, id_usuario), http.StatusNotFound)
		return
	}

	// Establezo el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados para el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(partido)
}

// Función para el handler PUT partido
func actualizarPartidoPorUsuario(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var request db.UpdatePartidoParams
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

}

// Funcion auxiliar de validación
func validarPartido(p PartidoRequest) {
	if *p.IDUsuario <= 0 || p.IDUsuario == nil {
		return errors.New("El partido no tiene ID de Usuario")
	}
	if p.Fecha == nil || p.Fecha.IsZero() {
		return errors.New("El partido no tiene fecha")
	}
	if p.Cancha == "" {
		return errors.New("El partido no tiene cancha")
	}
	if *p.Puntuacion <= 0 || p.Puntuacion == nil {
		return errors.New("El partido no tiene puntuacion")
	}
}

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

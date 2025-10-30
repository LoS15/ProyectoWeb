package main

import (
	db "ProyectoWeb/db/sqlc"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv" //Para la validacion de floats
	"strings" //Para la validacion de floats
	"time"
)

// Tipos auxiliares para las funciones de los handlers de Partido
type CrearPartidoCompletoRequest struct {
	IDUsuario          int32                              `json:"id_usuario"`
	Fecha              time.Time                          `json:"fecha"`
	Cancha             string                             `json:"cancha"`
	Puntuacion         int32                              `json:"puntuacion"`
	TipoEstadistica    string                             `json:"tipo_estadistica"`
	EstadisticaJugador *db.InsertEstadisticaJugadorParams `json:"estadistica_jugador,omitempty"`
	EstadisticaArquero *db.InsertEstadisticaArqueroParams `json:"estadistica_arquero,omitempty"`
}

type PartidoCompleto struct {
	IDUsuario          int32                              `json:"id_usuario"`
	IDPartido          int32                              `json:"id_partido"`
	Fecha              time.Time                          `json:"fecha"`
	Cancha             string                             `json:"cancha"`
	Puntuacion         int32                              `json:"puntuacion"`
	TipoEstadistica    string                             `json:"tipo_estadistica"`
	EstadisticaJugador *db.InsertEstadisticaJugadorParams `json:"estadistica_jugador,omitempty"`
	EstadisticaArquero *db.InsertEstadisticaArqueroParams `json:"estadistica_arquero,omitempty"`
}

// Tipos para poder hacer chequeos de nulos en tipos primitivos que no permiten control de nulos -- PARTIDO
// Para la inserción de partidos
type PartidoCrearRequest struct {
	IDUsuario  *int32     `json:"id_usuario"` // puntero para saber si es null
	Fecha      *time.Time `json:"fecha"`      // puntero para poder usar IsZero o nil
	Cancha     string     `json:"cancha"`     // string porque es obligatorio ("su forma de null" es estar vacio)
	Puntuacion *int32     `json:"puntuacion"` // puntero para saber si es null
}

// Para la actualización de partidos
type PartidoActualizacionRequest struct {
	IDUsuario  *int32     `json:"id_usuario"` // puntero para saber si es null
	IDPartido  *int32     `json:"id_partido"` // puntero para saber si es null
	Fecha      *time.Time `json:"fecha"`      // puntero para poder usar IsZero o nil
	Cancha     string     `json:"cancha"`     // string porque es obligatorio ("su forma de null" es estar vacio)
	Puntuacion *int32     `json:"puntuacion"` // puntero para saber si es null
}

// Funciones para handlers de Partido
// Funcion para el handler POST partido (solo crear partido) *(*solo para pruebas*)*
func crearPartido(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON
	var partido db.InsertPartidoParams
	err := json.NewDecoder(r.Body).Decode(&partido)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para crear el partido: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON para la creacion del partido
	partidoReq := PartidoCrearRequest{
		IDUsuario:  &partido.IDUsuario,
		Fecha:      &partido.Fecha,
		Cancha:     partido.Cancha,
		Puntuacion: &partido.Puntuacion,
	}
	err = validarCreacionPartido(partidoReq)
	if err != nil {
		// Si hay error en la validación del partido, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Inserto en la tabla Partido el nuevo partido
	nuevoPartido, err := queries.InsertPartido(ctx, partido)
	if err != nil {
		// Si ocurre un error al insertar el nuevo partido, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error creando partido: %v", err)
		http.Error(w, "Error creando partido", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con el partido como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(nuevoPartido)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON el partido creado: %v", err)
		http.Error(w, "Error codificando a JSON el partido creado", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler POST partido (para crear el partido completo con sus estadísticas, como realmente funcionaria)
func crearPartidoCompleto(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Uso transaccion por que son dos operaciones separadas en la BD y deben hacerse o ambas o ninguna
	// Comienzo la transaccion con la base de datos
	transaccion, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		// Si ocurre un error al iniciar la transaccion, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error iniciando la transaccion: %v", err)
		http.Error(w, "Error iniciando la transaccion", http.StatusInternalServerError)
		return
	}
	// Retrasamos el rollback hasta el final de la ejecución, en caso de que haya algún error
	defer transaccion.Rollback()

	// Creo el objeto que me permite ejecutar todas las consultas dentro de la transaccion
	qtran := queries.WithTx(transaccion)

	// Decodificar el JSON de la request en una entidad intermedia
	var request CrearPartidoCompletoRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para crear el partido completo junto a sus estadisticas: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Separo la informacion de la entidad intermedia
	// Valido los datos del JSON para la creacion del partido
	partidoReq := PartidoCrearRequest{
		IDUsuario:  &request.IDUsuario,
		Fecha:      &request.Fecha,
		Cancha:     request.Cancha,
		Puntuacion: &request.Puntuacion,
	}
	err = validarCreacionPartido(partidoReq)
	if err != nil {
		// Si hay error en la validación del partido, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Creo el partido para insertar
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
		log.Printf("Error creando partido: %v", err)
		http.Error(w, "Error creando partido", http.StatusInternalServerError)
		return
	}

	// Creo las estadisticas del partido, segun el tipo
	switch request.TipoEstadistica {
	case "jugador":
		// Si el tipo de estadisticas es de jugador

		if request.EstadisticaJugador == nil {
			// Si las estadisticas no están cargadas, lanzo código 400 y finalizo la ejecucion del handler
			log.Printf("Faltan datos de estadisticas de jugador: %v", err)
			http.Error(w, "Faltan datos de estadisticas de jugador", http.StatusBadRequest)
			return
		}

		// Valido los datos del JSON de las estadisticas de jugador
		estadisticasJugadorReq := EstadisticaJugadorRequest{
			Goles:            &request.EstadisticaJugador.Goles,
			Asistencias:      &request.EstadisticaJugador.Asistencias,
			PasesCompletados: request.EstadisticaJugador.PasesCompletados,
			DuelosGanados:    request.EstadisticaJugador.DuelosGanados,
		}
		err = validarEstadisticasJugador(estadisticasJugadorReq)
		if err != nil {
			// Si hay error en la validación de las estadisticas, lanzo código 400 y termino la ejecución del handler
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Creo las estadisticas de jugador para insertar
		datosNuevaEstadisticaJugador := db.InsertEstadisticaJugadorParams{
			IDUsuario:        nuevoPartido.IDUsuario,
			IDPartido:        nuevoPartido.IDPartido,
			Goles:            request.EstadisticaJugador.Goles,
			Asistencias:      request.EstadisticaJugador.Asistencias,
			PasesCompletados: request.EstadisticaJugador.PasesCompletados,
			DuelosGanados:    request.EstadisticaJugador.DuelosGanados,
		}

		// Inserto en la tabla EstadisticaJugador la nueva estadistica de jugador
		_, err = qtran.InsertEstadisticaJugador(ctx, datosNuevaEstadisticaJugador)
		if err != nil {
			// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
			log.Printf("Error creando estadisticas de jugador: %v", err)
			http.Error(w, "Error creando estadisticas de jugador", http.StatusInternalServerError)
			return
		}

	case "arquero":
		// Si el tipo de estadisticas es de arquero

		if request.EstadisticaArquero == nil {
			// Si las estadisticas no están cargadas, lanzo código 400 y finalizo la ejecucion del handler
			log.Printf("Faltan datos de estadisticas de arquero: %v", err)
			http.Error(w, "Faltan datos de estadisticas de arquero", http.StatusBadRequest)
			return
		}

		// Valido los datos del JSON de las estadisticas de arquero
		estadisticasArqueroReq := EstadisticaArqueroRequest{
			GolesRecibidos:    &request.EstadisticaArquero.GolesRecibidos,
			AtajadasClave:     &request.EstadisticaArquero.AtajadasClave,
			SaquesCompletados: request.EstadisticaArquero.SaquesCompletados,
		}
		err = validarEstadisticasArquero(estadisticasArqueroReq)
		if err != nil {
			// Si hay error en la validación de las estadisticas, lanzo código 400 y termino la ejecución del handler
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Creo las estadisticas de arquero para insertar
		datosNuevaEstadisticaArquero := db.InsertEstadisticaArqueroParams{
			IDUsuario:         nuevoPartido.IDUsuario,
			IDPartido:         nuevoPartido.IDPartido,
			GolesRecibidos:    request.EstadisticaArquero.GolesRecibidos,
			AtajadasClave:     request.EstadisticaArquero.AtajadasClave,
			SaquesCompletados: request.EstadisticaArquero.SaquesCompletados,
		}

		// Inserto en la tabla EstadisticaArquero la nueva estadistica de arquero
		_, err = qtran.InsertEstadisticaArquero(ctx, datosNuevaEstadisticaArquero)
		if err != nil {
			// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
			log.Printf("Error creando estadisticas de jugador: %v", err)
			http.Error(w, "Error creando estadisticas de arquero", http.StatusInternalServerError)
			return
		}
	default:
		// Si el tipo de estadisticas es otro, lanzo código 400 y finalizo la ejecucion del handler
		log.Printf("Error por el tipo de estadística no válido: %v", err)
		http.Error(w, "Tipo de estadística no válido", http.StatusBadRequest)
		return
	}

	// Confirmo la transaccion
	err = transaccion.Commit()
	if err != nil {
		// Si ocurre un error al confirmar la transaccion, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error confirmando la transaccion: %v", err)
		http.Error(w, "Error confirmando transacción", http.StatusInternalServerError)
		return
	}

	response := PartidoCompleto{
		IDUsuario:          nuevoPartido.IDUsuario,
		IDPartido:          nuevoPartido.IDPartido,
		Fecha:              nuevoPartido.Fecha,
		Cancha:             nuevoPartido.Cancha,
		Puntuacion:         nuevoPartido.Puntuacion,
		TipoEstadistica:    request.TipoEstadistica,
		EstadisticaJugador: request.EstadisticaJugador,
		EstadisticaArquero: request.EstadisticaArquero,
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con el partido y sus estadisticas como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON el partido completo creado junto a sus estadisticas: %v", err)
		http.Error(w, "Error codificando a JSON el partido completo creado junto a sus estadisticas", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET todos los partidos *(*solo para pruebas*)*
func listarTodosLosPartidos(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Obtengo todos los partidos de la tabla Partido
	partidos, err := queries.GetAllPartido(ctx)
	if err != nil {
		// Si ocurre un error obteniendo todos los partidos, lanzo código 404 y finalizo la ejecucion del handler
		log.Printf("Error obteniendo todos los partidos existentes: %v", err)
		http.Error(w, "Error obteniendo todos los partidos existentes", http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(partidos)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON todos los partidos existentes: %v", err)
		http.Error(w, "Error codificando a JSON todos los partidos existentes", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET todos los partidos para un usuario
func listarTodosLosPartidosPorUsuario(w http.ResponseWriter, r *http.Request, id_usuario int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Obtengo todos los partidos para un usuario dado de la tabla Partido
	partidos, err := queries.ListPartidosPorUsuario(ctx, id_usuario)
	if err != nil {
		// Si ocurre un error obteniendo todos los partidos para un usuario dado, lanzo código 404 y finalizo la ejecucion del handler
		log.Printf("Error obteniendo todos los partidos para el usuario %d: %v", id_usuario, err)
		http.Error(w, fmt.Sprintf("Error obteniendo todos los partidos para el usuario %d", id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados para el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(partidos)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON todos los partidos obtenidos para un usuario: %v", err)
		http.Error(w, "Error codificando a JSON todos los partidos obtenidos para un usuario", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET partido dado para un usuario dado
func obtenerPartidoPorUsuario(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
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
		log.Printf("Error obteniendo el partido %d para el usuario %d: %v", id_partido, id_usuario, err)
		http.Error(w, fmt.Sprintf("Error obteniendo el partido %d para el usuario %d", id_partido, id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con el partido para el usuario dado en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(partido)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON el partido obtenido: %v", err)
		http.Error(w, "Error codificando a JSON el partido obtenido", http.StatusInternalServerError)
		return
	}
}

// Función para el handler PUT partido
func actualizarPartido(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var partido db.UpdatePartidoParams
	err := json.NewDecoder(r.Body).Decode(&partido)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para actualizar el partido: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON para la actualizacion del partido
	partidoReq := PartidoActualizacionRequest{
		IDUsuario:  &partido.IDUsuario,
		IDPartido:  &partido.IDPartido,
		Fecha:      &partido.Fecha,
		Cancha:     partido.Cancha,
		Puntuacion: &partido.Puntuacion,
	}
	err = validarActualizarPartido(partidoReq)
	if err != nil {
		// Si hay error en la validación del partido, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Actualizo en la tabla Partido el partido
	nuevoPartido, err := queries.UpdatePartido(ctx, partido)
	if err != nil {
		// Si ocurre un error al actualizar el partido, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error actualizando partido: %v", err)
		http.Error(w, "Error actualizando partido", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 200 y respondo con el partido actualizado como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nuevoPartido)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON el partido actualizado: %v", err)
		http.Error(w, "Error codificando a JSON el partido actualizado", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler DELETE partido
func eliminarPartido(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Creo el objeto para eliminar el partido
	datosPartido := db.DeletePartidoParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Elimino de la tabla Partido el partido
	err := queries.DeletePartido(ctx, datosPartido)
	if err != nil {
		// Si ocurre un error al eliminar el partido, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error eliminando partido: %v", err)
		http.Error(w, "Error eliminando partido", http.StatusInternalServerError)
		return
	}

	// Lanzo el código 204 como respuesta exitosa de la operación
	w.WriteHeader(http.StatusNoContent)
}

// Tipo para poder hacer chequeo de nulos en tipos primitivos que no permiten control de nulos -- USUARIO
// Para la actualizacion de usuarios
type UsuarioActualizacionRequest struct {
	IDUsuario *int32 `json:"id_usuario"`
	Nombre    string `json:"nombre"`
	Apellido  string `json:"apellido"`
	Pais      string `json:"pais"`
}

// Funciones para handlers de Usuario
// Funcion para el handler POST Usuario
func crearUsuario(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var nuevoUsuario db.CreateUserParams
	err := json.NewDecoder(r.Body).Decode(&nuevoUsuario)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para crear el usuario: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON para la creación del usuario
	err = validarCreacionUsuario(nuevoUsuario)
	if err != nil {
		// Si hay error en la validación del usuario, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Inserto en la tabla Usuario el nuevo usuario
	usuario, err := queries.CreateUser(ctx, nuevoUsuario)
	if err != nil {
		// Si ocurre un error al insertar el nuevo usuario, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error creando usuario: %v", err)
		http.Error(w, "Error creando usuario", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con el usuario como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(usuario)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON el usuario creado: %v", err)
		http.Error(w, "Error codificando a JSON el usuario creado", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET todos los usuario *(*solo para pruebas*)*
func listarTodosLosUsuarios(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Obtengo todos los usuarios de la tabla Usuario
	usuarios, err := queries.GetAllUser(ctx)
	if err != nil {
		// Si ocurre un error obteniendo todos los usuarios, lanzo código 404 y finalizo la ejecucion del handler
		log.Printf("Error obteniendo todos los usuarios existentes: %v", err)
		http.Error(w, "Error obteniendo todos los usuarios existentes", http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los usuarios listados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(usuarios)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON todos los usuarios existentes: %v", err)
		http.Error(w, "Error codificando a JSON todos los usuarios existentes", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET para un usuario dado
func obtenerUsuario(w http.ResponseWriter, r *http.Request, id_usuario int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Obtengo un usuario dado
	usuario, err := queries.GetUsuario(ctx, id_usuario)
	if err != nil {
		// Si ocurre un error obteniendo el usuario dado, lanzo código 404 y finalizo la ejecucion del handler
		log.Printf("Error obteniendo el usuario %d: %v", id_usuario, err)
		http.Error(w, fmt.Sprintf("Error obteniendo el usuario %d", id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(usuario)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON el usuario obtenido: %v", err)
		http.Error(w, "Error codificando a JSON el usuario obtenido", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler PUT usuario
func actualizarUsuario(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var usuario db.UpdateUserParams
	err := json.NewDecoder(r.Body).Decode(&usuario)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para actualizar el usuario: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON para la actualizacion del usuario
	usuarioReq := UsuarioActualizacionRequest{
		IDUsuario: &usuario.IDUsuario,
		Nombre:    usuario.Nombre,
		Apellido:  usuario.Apellido,
		Pais:      usuario.Pais,
	}
	err = validarActualizarUsuario(usuarioReq)
	if err != nil {
		// Si hay error en la validación del usuario, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Actualizo en la tabla Usuario el usuario
	nuevoUsuario, err := queries.UpdateUser(ctx, usuario)
	if err != nil {
		// Si ocurre un error al actualizar el usuario, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error actualizando usuario: %v", err)
		http.Error(w, "Error actualizando usuario", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 200 y respondo con el usuario actualizado como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nuevoUsuario)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON el usuario actualizado: %v", err)
		http.Error(w, "Error codificando a JSON el usuario actualizado", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler DELETE usuario
func eliminarUsuario(w http.ResponseWriter, r *http.Request, id_usuario int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Elimino en la tabla Partido el partido
	err := queries.DeleteUsuario(ctx, id_usuario)
	if err != nil {
		// Si ocurre un error al eliminar el usuario, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error eliminando usuario: %v", err)
		http.Error(w, "Error eliminando usuario", http.StatusInternalServerError)
		return
	}

	// Lanzo el código 204 como respuesta exitosa de la operación
	w.WriteHeader(http.StatusNoContent)
}

// Tipo para poder hacer chequeo de nulos en tipos primitivos que no permiten control de nulos -- ESTADISTICAS AMBAS
// Para la insercion de estadisticas *(*solo para pruebas*)*
type EstadisticaCrearJugadorRequest struct {
	IDUsuario        *int32 `json:"id_usuario"`        // puntero para saber si es null
	IDPartido        *int32 `json:"id_partido"`        // puntero para saber si es null
	Goles            *int32 `json:"goles"`             // puntero para saber si vino
	Asistencias      *int32 `json:"asistencias"`       // puntero para saber si vino
	PasesCompletados string `json:"pases_completados"` // string porque es obligatorio ("su forma de null" es estar vacio)
	DuelosGanados    string `json:"duelos_ganados"`    // string porque es obligatorio ("su forma de null" es estar vacio)
}

type EstadisticaCrearArqueroRequest struct {
	IDUsuario         *int32 `json:"id_usuario"`         // puntero para saber si es null
	IDPartido         *int32 `json:"id_partido"`         // puntero para saber si es null
	GolesRecibidos    *int32 `json:"goles_recibidos"`    // puntero para saber si vino
	AtajadasClave     *int32 `json:"atajadas_clave"`     // puntero para saber si vino
	SaquesCompletados string `json:"saques_completados"` // string porque es obligatorio ("su forma de null" es estar vacio)
}

// Para retornar una estadistica independientemente de su tipo
type RetornoEstadistica struct {
	TipoEstadistica    string                 `json:"tipo_estadistica"`
	EstadisticaJugador *db.EstadisticaJugador `json:"estadistica_jugador,omitempty"`
	EstadisticaArquero *db.EstadisticaArquero `json:"estadistica_arquero,omitempty"`
}

// Para la actualizacion e insercion (en el flujo de insercion de partido completo) de estadisticas
type EstadisticaJugadorRequest struct {
	Goles            *int32 `json:"goles"`             // puntero para saber si vino
	Asistencias      *int32 `json:"asistencias"`       // puntero para saber si vino
	PasesCompletados string `json:"pases_completados"` // string porque es obligatorio ("su forma de null" es estar vacio)
	DuelosGanados    string `json:"duelos_ganados"`    // string porque es obligatorio ("su forma de null" es estar vacio)
}

type EstadisticaArqueroRequest struct {
	GolesRecibidos    *int32 `json:"goles_recibidos"`    // puntero para saber si vino
	AtajadasClave     *int32 `json:"atajadas_clave"`     // puntero para saber si vino
	SaquesCompletados string `json:"saques_completados"` // string porque es obligatorio ("su forma de null" es estar vacio)
}

// Funciones para handlers de EstadisticasJugador
// Funcion para el handler POST EstadisticaJugador *(*solo para pruebas*)*
func crearEstadisticasJugador(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var estadistica db.InsertEstadisticaJugadorParams
	err := json.NewDecoder(r.Body).Decode(&estadistica)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para crear la estadistica de jugador: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON de las estadisticas de jugador
	estadisticasJugadorReq := EstadisticaCrearJugadorRequest{
		IDUsuario:        &estadistica.IDUsuario,
		IDPartido:        &estadistica.IDPartido,
		Goles:            &estadistica.Goles,
		Asistencias:      &estadistica.Asistencias,
		PasesCompletados: estadistica.PasesCompletados,
		DuelosGanados:    estadistica.DuelosGanados,
	}
	err = validarCrearEstadisticasJugador(estadisticasJugadorReq)
	if err != nil {
		// Si hay error en la validación de las estadisticas, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Inserto en la tabla EstadisticaJugador la nueva estadistica de jugador
	nuevaEstadistica, err := queries.InsertEstadisticaJugador(ctx, estadistica)
	if err != nil {
		// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error creando estadisticas de jugador: %v", err)
		http.Error(w, "Error creando estadisticas de jugador", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con las estadisticas de jugador como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(nuevaEstadistica)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON la estadistica de jugador creada: %v", err)
		http.Error(w, "Error codificando a JSON la estadistica de jugador creada", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET EstadisticaJugador para un partido y usuario dado
func obtenerEstadisticasJugador(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Crea el tipo auxiliar necesario para la operación sqlc
	informacionEstadistica := db.GetEstadisticaJugadorParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Obtengo una estadistica para un partido y un usuario dado
	estadistica, err := queries.GetEstadisticaJugador(ctx, informacionEstadistica)
	if err != nil {
		// Si ocurre un error obteniendo la estadistica para un partido y usuario dado, lanzo código 404 y finalizo la ejecucion del handler
		log.Printf("Error obteniendo la estadistica de jugador para el partido %d y el usuario %d: %v", id_partido, id_usuario, err)
		http.Error(w, fmt.Sprintf("Error obteniendo la estadistica de jugador para el partido %d y el usuario %d", id_partido, id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados para el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(estadistica)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON la estadistica de jugador obtenida: %v", err)
		http.Error(w, "Error codificando a JSON la estadistica de jugador obtenida", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler PUT EstadisticaJugador
func actualizarEstadisticasJugador(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var estadistica db.UpdateEstadisticaJugadorParams
	err := json.NewDecoder(r.Body).Decode(&estadistica)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para actualizar la estadistica de jugador: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON de las estadisticas de jugador
	estadisticasJugadorReq := EstadisticaJugadorRequest{
		Goles:            &estadistica.Goles,
		Asistencias:      &estadistica.Asistencias,
		PasesCompletados: estadistica.PasesCompletados,
		DuelosGanados:    estadistica.DuelosGanados,
	}
	err = validarEstadisticasJugador(estadisticasJugadorReq)
	if err != nil {
		// Si hay error en la validación de las estadisticas, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Inserto en la tabla EstadisticaJugador la estadistica actualizada de jugador
	nuevaEstadistica, err := queries.UpdateEstadisticaJugador(ctx, estadistica)
	if err != nil {
		// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error actualizando estadisticas de jugador: %v", err)
		http.Error(w, "Error actualizando estadisticas de jugador", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con las estadisticas de jugador como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(nuevaEstadistica)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON la estadistica de jugador actualizada: %v", err)
		http.Error(w, "Error codificando a JSON la estadistica de jugador actualizada", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler DELETE EstadisticasJugador *(*solo para pruebas*)*
func eliminarEstadisticaJugador(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Creo el objeto para eliminar la estadistica de jugador
	datosEstadisticaJugador := db.DeleteEstadisticaJugadorParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Elimino de la tabla EstadisticaArquero la estadistica
	err := queries.DeleteEstadisticaJugador(ctx, datosEstadisticaJugador)
	if err != nil {
		// Si ocurre un error al eliminar la estadistica, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error eliminando estadisticas de jugador: %v", err)
		http.Error(w, "Error eliminando estadistica de jugador", http.StatusInternalServerError)
		return
	}

	// Lanzo el código 204 como respuesta exitosa de la operación
	w.WriteHeader(http.StatusNoContent)
}

// Funciones para handlers de EstadisticasArquero
// Funcion para el handler POST EstadisticaArquero *(*solo para pruebas*)*
func crearEstadisticasArquero(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var estadistica db.InsertEstadisticaArqueroParams
	err := json.NewDecoder(r.Body).Decode(&estadistica)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para crear la estadistica de arquero: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON de las estadisticas, en este caso de jugador
	estadisticasArqueroReq := EstadisticaCrearArqueroRequest{
		IDUsuario:         &estadistica.IDUsuario,
		IDPartido:         &estadistica.IDPartido,
		GolesRecibidos:    &estadistica.GolesRecibidos,
		AtajadasClave:     &estadistica.AtajadasClave,
		SaquesCompletados: estadistica.SaquesCompletados,
	}
	err = validarCrearEstadisticasArquero(estadisticasArqueroReq)
	if err != nil {
		// Si hay error en la validación de las estadisticas, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Inserto en la tabla EstadisticaArquero la nueva estadistica de arquero
	nuevaEstadistica, err := queries.InsertEstadisticaArquero(ctx, estadistica)
	if err != nil {
		// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error creando estadisticas de arquero: %v", err)
		http.Error(w, "Error creando estadisticas de arquero", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con las estadisticas de arquero como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(nuevaEstadistica)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON la estadistica de arquero creada: %v", err)
		http.Error(w, "Error codificando a JSON la estadistica de arquero creada", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET EstadisticaJugador para un partido y usuario dado
func obtenerEstadisticasArquero(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Crea el tipo auxiliar necesario para la operación sqlc
	informacionEstadistica := db.GetEstadisticaArqueroParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Obtengo una estadistica para un partido y un usuario dado
	estadistica, err := queries.GetEstadisticaArquero(ctx, informacionEstadistica)
	if err != nil {
		// Si ocurre un error obteniendo la estadistica para un partido y usuario dado, lanzo código 404 y finalizo la ejecucion del handler
		log.Printf("Error obteniendo la estadistica de arquero para el partido %d y el usuario %d: %v", id_partido, id_usuario, err)
		http.Error(w, fmt.Sprintf("Error obteniendo la estadistica de arquero para el partido %d y el usuario %d", id_partido, id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados para el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(estadistica)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON la estadistica de arquero obtenida: %v", err)
		http.Error(w, "Error codificando a JSON la estadistica de arquero obtenida", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler PUT EstadisticaJugador
func actualizarEstadisticasArquero(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON de la request
	var estadistica db.UpdateEstadisticaArqueroParams
	err := json.NewDecoder(r.Body).Decode(&estadistica)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		log.Printf("Error al decodificar el JSON para actualizar la estadistica de arquero: %v", err)
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

	// Valido los datos del JSON de las estadisticas de jugador
	estadisticasArqueroReq := EstadisticaArqueroRequest{
		GolesRecibidos:    &estadistica.GolesRecibidos,
		AtajadasClave:     &estadistica.AtajadasClave,
		SaquesCompletados: estadistica.SaquesCompletados,
	}
	err = validarEstadisticasArquero(estadisticasArqueroReq)
	if err != nil {
		// Si hay error en la validación de las estadisticas, lanzo código 400 y termino la ejecución del handler
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Inserto en la tabla EstadisticaJugador la estadistica actualizada de jugador
	nuevaEstadistica, err := queries.UpdateEstadisticaArquero(ctx, estadistica)
	if err != nil {
		// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error actualizando estadisticas de arquero: %v", err)
		http.Error(w, "Error actualizando estadisticas de arquero", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con las estadisticas de jugador como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(nuevaEstadistica)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON la estadistica de arquero actualizada: %v", err)
		http.Error(w, "Error codificando a JSON la estadistica de arquero actualizada", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler DELETE EstadisticasArquero *(*solo para pruebas*)*
func eliminarEstadisticasArquero(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Creo el objeto para eliminar la estadistica de arquero
	datosEstadisticaArquero := db.DeleteEstadisticaArqueroParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Elimino de la tabla EstadisticaArquero la estadistica
	err := queries.DeleteEstadisticaArquero(ctx, datosEstadisticaArquero)
	if err != nil {
		// Si ocurre un error al eliminar la estadistica, lanzo código 500 y finalizo la ejecucion del handler
		log.Printf("Error eliminando estadisticas de arquero: %v", err)
		http.Error(w, "Error eliminando estadisticas de arquero", http.StatusInternalServerError)
		return
	}

	// Lanzo el código 204 como respuesta exitosa de la operación
	w.WriteHeader(http.StatusNoContent)
}

// Funcion para el handler GET Estadisticas para un partido y usuario dado
func obtenerEstadisticas(w http.ResponseWriter, r *http.Request, id_usuario int32, id_partido int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Crea el tipo auxiliar necesario para la operación sqlc
	informacionEstadisticaJugador := db.GetEstadisticaJugadorParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Crea el tipo auxiliar necesario para la operación sqlc
	informacionEstadisticaArquero := db.GetEstadisticaArqueroParams{
		IDUsuario: id_usuario,
		IDPartido: id_partido,
	}

	// Creo las variables que son punteros a las estadisticas encontradas para luego usarlas
	var estadisticaJugadorPTR *db.EstadisticaJugador
	var estadisticaArqueroPTR *db.EstadisticaArquero

	// Obtengo una estadistica de jugador para un partido y un usuario dado, y la guardo en una variable que es puntero a la misma para poder preguntar por nil
	estadisticaJugador, err := queries.GetEstadisticaJugador(ctx, informacionEstadisticaJugador)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// Otro tipo de error (había estadistica de jugador) al obtener la estadistica, lanzo código 500 y termino la ejecución
			log.Printf("Error interno al obtener estadística de jugador: %v", err)
			http.Error(w, "Error interno al obtener estadística de jugador", http.StatusInternalServerError)
			return
		}
	} else {
		estadisticaJugadorPTR = &estadisticaJugador
	}

	// Obtengo una estadistica para un partido y un usuario dado, y la guardo en una variable que es puntero a la misma para poder preguntar por nil
	estadisticaArquero, err := queries.GetEstadisticaArquero(ctx, informacionEstadisticaArquero)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// Otro tipo de error (había estadistica de arquero) al obtener la estadistica, lanzo código 500 y termino la ejecución
			log.Printf("Error interno al obtener estadística de arquero: %v", err)
			http.Error(w, "Error interno al obtener estadística de arquero", http.StatusInternalServerError)
			return
		}
	} else {
		estadisticaArqueroPTR = &estadisticaArquero
	}

	// Creo la variable que retornare en caso de que exista la estadistica
	estadistica := RetornoEstadistica{}

	// Consulto si existe la estadistica de algún tipo o si no existe ninguna (en este caso, lanzo código 404 y corto la ejecución)
	if estadisticaJugadorPTR == nil && estadisticaArqueroPTR == nil {
		// No existe la estadistica buscada, de ninguno de los dos tipos
		// Si ocurre un error obteniendo la estadistica para un partido y usuario dado, lanzo código 404 y finalizo la ejecucion del handler
		log.Printf("Error no existen estadisticas de ningun tipo para el partido %d y el usuario %d: %v", id_partido, id_usuario, err)
		http.Error(w, fmt.Sprintf("Error no existen estadisticas de ningun tipo para el partido %d y el usuario %d", id_partido, id_usuario), http.StatusNotFound)
		return

	} else if estadisticaJugadorPTR == nil {
		// Existe estadistica de arquero
		estadistica.TipoEstadistica = "arquero"
		estadistica.EstadisticaJugador = estadisticaJugadorPTR
		estadistica.EstadisticaArquero = estadisticaArqueroPTR

	} else if estadisticaArqueroPTR == nil {
		// Existe estadistica de jugador
		estadistica.TipoEstadistica = "jugador"
		estadistica.EstadisticaJugador = estadisticaJugadorPTR
		estadistica.EstadisticaArquero = estadisticaArqueroPTR

	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con la estadistica encontrada en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Printf("%+v", estadistica)
	err = json.NewEncoder(w).Encode(estadistica)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		log.Printf("Error codificando a JSON la estadistica obtenida: %v", err)
		http.Error(w, "Error codificando a JSON la estadistica obtenida", http.StatusInternalServerError)
		return
	}
}

// Funciones auxiliares de validación
func validarCreacionPartido(partido PartidoCrearRequest) error {
	if partido.IDUsuario == nil || *partido.IDUsuario <= 0 {
		return errors.New("El partido no tiene el dato obligatorio: IDUsuario")
	}
	if partido.Fecha == nil || partido.Fecha.IsZero() {
		return errors.New("El partido no tiene el dato obligatorio: Fecha")
	}
	if partido.Cancha == "" {
		return errors.New("El partido no tiene el dato obligatorio: Cancha")
	}
	if partido.Puntuacion == nil || *partido.Puntuacion < 0 {
		return errors.New("El partido no tiene el dato obligatorio: Puntuacion")
	}

	// Si pasa la validación
	return nil
}

func validarActualizarPartido(partido PartidoActualizacionRequest) error {
	if partido.IDUsuario == nil || *partido.IDUsuario <= 0 {
		return errors.New("El partido no tiene el dato obligatorio: IDUsuario")
	}
	if partido.IDPartido == nil || *partido.IDPartido <= 0 {
		return errors.New("El partido no tiene el dato obligatorio: IDPartido")
	}
	if partido.Fecha == nil || partido.Fecha.IsZero() {
		return errors.New("El partido no tiene el dato obligatorio: Fecha")
	}
	if partido.Cancha == "" {
		return errors.New("El partido no tiene el dato obligatorio: Cancha")
	}
	if partido.Puntuacion == nil || *partido.Puntuacion < 0 {
		return errors.New("El partido no tiene el dato obligatorio: Puntuacion")
	}

	// Si pasa la validación
	return nil
}

func validarCrearEstadisticasJugador(estadisticas EstadisticaCrearJugadorRequest) error {
	if estadisticas.IDUsuario == nil || *estadisticas.IDUsuario <= 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: IDUsuario")
	}
	if estadisticas.IDPartido == nil || *estadisticas.IDPartido <= 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: IDPartido")
	}
	if estadisticas.Goles == nil || *estadisticas.Goles < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Goles")
	}
	if estadisticas.Asistencias == nil || *estadisticas.Asistencias < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Asistencias")
	}
	if estadisticas.PasesCompletados != "" {
		// Controlo solo para los valores No Nulo
		//Valido el formato y valor
		err := validarFormatoPorcentaje(&estadisticas.PasesCompletados, "pases completados")
		if err != nil {
			return err
		}
	}
	if estadisticas.DuelosGanados != "" {
		// Controlo solo para los valores No Nulo
		//Valido el formato y valor
		err := validarFormatoPorcentaje(&estadisticas.DuelosGanados, "duelos ganados")
		if err != nil {
			return err
		}
	}

	// Si pasa la validación
	return nil
}

func validarCrearEstadisticasArquero(estadisticas EstadisticaCrearArqueroRequest) error {
	if estadisticas.IDUsuario == nil || *estadisticas.IDUsuario <= 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: IDUsuario")
	}
	if estadisticas.IDPartido == nil || *estadisticas.IDPartido <= 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: IDPartido")
	}
	if estadisticas.GolesRecibidos == nil || *estadisticas.GolesRecibidos < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Goles Recibidos")
	}
	if estadisticas.AtajadasClave == nil || *estadisticas.AtajadasClave < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Atajadas Clave")
	}
	if estadisticas.SaquesCompletados != "" {
		// Controlo solo para los valores No Nulos
		//Valido el formato y valor
		err := validarFormatoPorcentaje(&estadisticas.SaquesCompletados, "saques completados")
		if err != nil {
			return err
		}
	}

	// Si pasa la validación
	return nil
}

func validarEstadisticasJugador(estadisticas EstadisticaJugadorRequest) error {
	if estadisticas.Goles == nil || *estadisticas.Goles < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Goles")
	}
	if estadisticas.Asistencias == nil || *estadisticas.Asistencias < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Asistencias")
	}
	if estadisticas.PasesCompletados != "" {
		// Controlo solo para los valores No Nulo
		//Valido el formato y valor
		err := validarFormatoPorcentaje(&estadisticas.PasesCompletados, "pases completados")
		if err != nil {
			return err
		}
	}
	if estadisticas.DuelosGanados != "" {
		// Controlo solo para los valores No Nulo
		//Valido el formato y valor
		err := validarFormatoPorcentaje(&estadisticas.DuelosGanados, "duelos ganados")
		if err != nil {
			return err
		}
	}

	// Si pasa la validación
	return nil
}

func validarEstadisticasArquero(estadisticas EstadisticaArqueroRequest) error {
	if estadisticas.GolesRecibidos == nil || *estadisticas.GolesRecibidos < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Goles Recibidos")
	}
	if estadisticas.AtajadasClave == nil || *estadisticas.AtajadasClave < 0 {
		return errors.New("La estadistica no tiene el dato obligatorio: Atajadas Clave")
	}
	if estadisticas.SaquesCompletados != "" {
		// Controlo solo para los valores No Nulos
		//Valido el formato y valor
		err := validarFormatoPorcentaje(&estadisticas.SaquesCompletados, "saques completados")
		if err != nil {
			return err
		}
	}

	// Si pasa la validación
	return nil
}

func validarFormatoPorcentaje(dato *string, atributo string) error {
	// Reemplazo la coma por el punto, en caso de que el formato no fuera el adecuado
	*dato = strings.ReplaceAll(*dato, ",", ".")

	// Creo variable local que uso para validar
	s := *dato

	if s == "" {
		//Si el dato tiene un valor vacío
		return fmt.Errorf("El dato %s es vacío, debe ser un porcentaje entre 0.00 y 1.00 (hasta dos decimales)", atributo)
	} else {
		// Si el dato tiene un valor no vacío

		// Parseo el string como float
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			// Si ocurre un error en el parseo
			return fmt.Errorf("El dato %s es inválido, debe ser un porcentaje entre 0.00 y 1.00 (hasta dos decimales)", atributo)
		}

		// Valido el valor para que este dentro del rango de valores
		if val < 0.00 || val > 1.00 {
			// Si el valor es inválido
			return fmt.Errorf("El dato %s es inválido, debe ser un porcentaje entre 0.00 y 1.00 (hasta dos decimales)", atributo)
		}

		// Valido el formato de float que permite la BD (decimal(3,2))
		parts := strings.Split(s, ".")
		if len(parts) == 2 && len(parts[1]) > 2 {
			// Si existe el punto y por ende, la parte decimal (primer condición), y tiene mas de dos decimales (segunda condición)
			return fmt.Errorf("El dato %s es inválido, debe ser un porcentaje entre 0.00 y 1.00 (hasta dos decimales)", atributo)
		}

		// Si pasa la validación
		return nil
	}
}

func validarCreacionUsuario(usuario db.CreateUserParams) error {
	if usuario.Nombre == "" {
		return errors.New("El usuario no tiene el dato obligatorio: Nombre")
	}
	if usuario.Apellido == "" {
		return errors.New("El usuario no tiene el dato obligatorio: Apellido")
	}
	if usuario.Pais == "" {
		return errors.New("El usuario no tiene el dato obligatorio: Pais")
	}

	// Si pasa la validación
	return nil
}

func validarActualizarUsuario(usuario UsuarioActualizacionRequest) error {
	if usuario.IDUsuario == nil || *usuario.IDUsuario <= 0 {
		return errors.New("El usuario no tiene el dato obligatorio: IDUsuario")
	}
	if usuario.Nombre == "" {
		return errors.New("El usuario no tiene el dato obligatorio: Nombre")
	}
	if usuario.Apellido == "" {
		return errors.New("El usuario no tiene el dato obligatorio: Apellido")
	}
	if usuario.Pais == "" {
		return errors.New("El usuario no tiene el dato obligatorio: Pais")
	}

	// Si pasa la validación
	return nil
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, "static/index.html")
}

func ResourceNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 - Recurso no encontrado")
}

func UsuariosHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// Caso: /usuarios/{id}
	if len(pathParts) == 2 && pathParts[0] == "usuarios" {
		id, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			obtenerUsuario(w, r, int32(id))
		case "DELETE":
			eliminarUsuario(w, r, int32(id))
		default:
			http.Error(w, "Método no permitido para esta ruta", http.StatusMethodNotAllowed)
		}
		return
	}

	// Caso: /usuarios (sin ID)
	if len(pathParts) == 1 && pathParts[0] == "usuarios" {
		switch r.Method {
		case "POST":
			crearUsuario(w, r)
		case "PUT":
			actualizarUsuario(w, r)
		case "GET":
			listarTodosLosUsuarios(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
		return
	}

	// Si no coincide con ninguna forma válida (/usuarios o /usuarios/{id})
	http.NotFound(w, r)
}

func PartidosHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// Caso: /usuarios/{id}
	if len(pathParts) == 3 && pathParts[0] == "partidos" {
		id_usuario, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
		id_partido, err := strconv.Atoi(pathParts[2])
		switch r.Method {
		case "GET":
			obtenerPartidoPorUsuario(w, r, int32(id_usuario), int32(id_partido))
		case "DELETE":
			eliminarPartido(w, r, int32(id_usuario), int32(id_partido))
		default:
			http.Error(w, "Método no permitido para esta ruta", http.StatusMethodNotAllowed)
		}
		return
	}
	// Caso: /usuarios (sin ID)
	if len(pathParts) == 1 && pathParts[0] == "partidos" {
		switch r.Method {
		case "POST":
			crearPartido(w, r)
		case "PUT":
			actualizarPartido(w, r)
		case "GET":
			listarTodosLosPartidos(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
		return
	}
}

// funcion de ayuda para generar correctamente los idUsuario e idJugador que se utilizan para eliminar u obtener partidos
func funcionConversionIdUsuarioAndIdJugador(w http.ResponseWriter, r *http.Request, funcion func(http.ResponseWriter, *http.Request, int32, int32)) {
	idUsuario, err := strconv.ParseInt(r.URL.Query().Get("id_usuario"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	idPartido, err := strconv.ParseInt(r.URL.Query().Get("id_partido"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	idUsuario32 := int32(idUsuario)
	idPartido32 := int32(idPartido)
	funcion(w, r, idUsuario32, idPartido32)
}

func funcionConversionIdUsuario(w http.ResponseWriter, r *http.Request, funcion func(http.ResponseWriter, *http.Request, int32)) {
	idUsuario, err := strconv.ParseInt(r.URL.Query().Get("id_usuario"), 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	idUsuario32 := int32(idUsuario)
	funcion(w, r, idUsuario32)
}

func EstadisticasJugadorHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Caso: /estadistica-jugador/id_usuario/id_partido
	if len(pathParts) == 3 && pathParts[0] == "estadisticas-jugador" {
		id_usuario, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
		id_partido, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case "GET":
			obtenerEstadisticasJugador(w, r, int32(id_usuario), int32(id_partido))
		case "DELETE":
			eliminarEstadisticaJugador(w, r, int32(id_usuario), int32(id_partido))
		default:
			http.Error(w, "Método no permitido para esta ruta", http.StatusMethodNotAllowed)
		}
		return
	}
	// Caso: /usuarios (sin ID)
	if len(pathParts) == 1 && pathParts[0] == "estadisticas-jugador" {
		switch r.Method {
		case "POST":
			crearEstadisticasJugador(w, r)
		case "PUT":
			actualizarEstadisticasJugador(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
		return
	}
}

func EstadisticasArqueroHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Caso: /estadistica-jugador/id_usuario/id_partido
	if len(pathParts) == 3 && pathParts[0] == "estadisticas-arquero" {
		id_usuario, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
		id_partido, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case "GET":
			obtenerEstadisticasArquero(w, r, int32(id_usuario), int32(id_partido))
		case "DELETE":
			eliminarEstadisticasArquero(w, r, int32(id_usuario), int32(id_partido))
		default:
			http.Error(w, "Método no permitido para esta ruta", http.StatusMethodNotAllowed)
		}
		return
	}
	// Caso: /usuarios (sin ID)
	if len(pathParts) == 1 && pathParts[0] == "estadisticas-arquero" {
		switch r.Method {
		case "POST":
			crearEstadisticasArquero(w, r)
		case "PUT":
			actualizarEstadisticasArquero(w, r)
		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		}
		return
	}
}

func obtenerEstadisticasHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Caso: /estadisticas/id_usuario/id_partido
	if len(pathParts) == 3 && pathParts[0] == "estadisticas" {
		id_usuario, err := strconv.Atoi(pathParts[1])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}
		id_partido, err := strconv.Atoi(pathParts[2])
		if err != nil {
			http.Error(w, "ID invalido", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case "GET":
			obtenerEstadisticas(w, r, int32(id_usuario), int32(id_partido))
			return
		default:
			http.Error(w, "Método no permitido para esta ruta", http.StatusMethodNotAllowed)
			return
		}
	}
	http.Error(w, "Método no permitido para esta ruta", http.StatusMethodNotAllowed)
	return
}

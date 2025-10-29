package main

import (
	db "ProyectoWeb/db/sqlc"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	EstadisticaJugador *db.InsertEstadisticaJugadorParams `json:"estadisticas_jugador,omitempty"`
	EstadisticaArquero *db.InsertEstadisticaArqueroParams `json:"estadistica_arquero,omitempty"`
}

// Tipos para poder hacer chequeos de nulos en tipos primitivos que no permiten control de nulos
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

// Para la inserción de estadisticas
type EstadisticaJugadorRequest struct {
	Goles            *int32         `json:"goles"`                       // puntero para saber si vino
	Asistencias      *int32         `json:"asistencias"`                 // puntero para saber si vino
	PasesCompletados sql.NullString `json:"pases_completados,omitempty"` // sql.NullString porque es opcional ("su forma de null" es estar vacio o ser null)
	DuelosGanados    sql.NullString `json:"duelos_ganados,omitempty"`    // sql.NullString porque es opcional ("su forma de null" es estar vacio o ser null)
}

// Para la inserción de estadisticas
type EstadisticaArqueroRequest struct {
	GolesRecibidos    *int32         `json:"goles_recibidos"`    // puntero para saber si vino
	AtajadasClave     *int32         `json:"atajadas_clave"`     // puntero para saber si vino
	SaquesCompletados sql.NullString `json:"saques_completados"` // sql.NullString porque es opcional ("su forma de null" es estar vacio o ser null)
}

// Funciones para handlers de Partido
// Funcion para el handler POST partido (solo crear partido, para pruebas únicamente)
func crearPartido(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Decodificar el JSON para la inserción del partido
	var request CrearPartidoCompletoRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// Si ocurre un error al decodificar el JSON, lanzo un código 400 y finalizo la ejecucion del handler
		http.Error(w, "Los datos envíados son inválidos", http.StatusBadRequest)
		return
	}

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
	nuevoPartido, err := queries.InsertPartido(ctx, datosNuevoPartido)
	if err != nil {
		// Si ocurre un error al insertar el nuevo partido, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error creando partido", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con el partido como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(nuevoPartido)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
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

		// Valido los datos del JSON de las estadisticas, en este caso de jugador
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

		_, err = qtran.InsertEstadisticaJugador(ctx, datosNuevaEstadisticaJugador)
		if err != nil {
			// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
			http.Error(w, "Error creando estadisticas de jugador", http.StatusInternalServerError)
			return
		}

	case "arquero":
		// Si el tipo de estadisticas es de arquero

		if request.EstadisticaArquero == nil {
			// Si las estadisticas no están cargadas, lanzo código 400 y finalizo la ejecucion del handler
			http.Error(w, "Faltan datos de estadisticas de arquero", http.StatusBadRequest)
			return
		}

		// Valido los datos del JSON de las estadisticas, en este caso de arquero
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

		_, err = qtran.InsertEstadisticaArquero(ctx, datosNuevaEstadisticaArquero)
		if err != nil {
			// Si ocurre un error al insertar las nuevas estadisticas, lanzo código 500 y finalizo la ejecucion del handler
			http.Error(w, "Error creando estadisticas de arquero", http.StatusInternalServerError)
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

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con el partido y sus estadisticas como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(request)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error codificando a JSON el partido creado junto a sus estadisticas", http.StatusInternalServerError)
		return
	}
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

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(partidos)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error codificando a JSON todos los partidos obtenidos", http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error obteniendo todos los partidos para el usuario %d", id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados para el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(partidos)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error codificando a JSON todos los partidos obtenidos", http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error obteniendo el partido %d para el usuario %d", id_partido, id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los partidos listados para el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(partido)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
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

	// Creo el objeto para actualizar el partido
	datosPartido := db.UpdatePartidoParams{
		IDUsuario:  partido.IDUsuario,
		IDPartido:  partido.IDPartido,
		Fecha:      partido.Fecha,
		Cancha:     partido.Cancha,
		Puntuacion: partido.Puntuacion,
	}

	// Actualizo el partido
	nuevoPartido, err := queries.UpdatePartido(ctx, datosPartido)
	if err != nil {
		// Si ocurre un error al actualizar el partido, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error actualizando partido", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 200 y respondo con el partido actualizado como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nuevoPartido)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
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

	// Elimino el partido
	err := queries.DeletePartido(ctx, datosPartido)
	if err != nil {
		// Si ocurre un error al eliminar el partido, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error eliminando partido", http.StatusInternalServerError)
		return
	}

	// Lanzo el código 204 como respuesta exitosa de la operación
	w.WriteHeader(http.StatusNoContent)
}

// Tipo para poder hacer chequeo de nulos en tipos primitivos que no permiten control de nulos
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

	// Creo el usuario
	usuario, err := queries.CreateUser(ctx, nuevoUsuario)
	if err != nil {
		// Si ocurre un error al insertar el nuevo usuario, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error creando usuario", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 201 y respondo con el usuario como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(usuario)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error codificando a JSON el usuario creado", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler GET todos los usuario
func listarTodosLosUsuarios(w http.ResponseWriter, r *http.Request) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Obtengo todos los usuarios de la tabla Usuario
	usuarios, err := queries.GetAllUser(ctx)
	if err != nil {
		// Si ocurre un error obteniendo todos los usuarios, lanzo código 404 y finalizo la ejecucion del handler
		http.Error(w, "Error obteniendo todos los usuarios existentes", http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con los usuarios listados en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(usuarios)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error codificando a JSON todos los usuarios obtenidos", http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error obteniendo el usuario %d", id_usuario), http.StatusNotFound)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo código 200 y respondo con el usuario en formato JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(usuario)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
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

	// Creo el objeto para actualizar el usuario
	datosUsuario := db.UpdateUserParams{
		IDUsuario: usuario.IDUsuario,
		Nombre:    usuario.Nombre,
		Apellido:  usuario.Apellido,
		Pais:      usuario.Pais,
	}

	// Actualizo el usuario
	nuevoUsuario, err := queries.UpdateUser(ctx, datosUsuario)
	if err != nil {
		// Si ocurre un error al actualizar el usuario, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error actualizando usuario", http.StatusInternalServerError)
		return
	}

	// Establezco el header de respuesta de tipo JSON, lanzo el código 200 y respondo con el usuario actualizado como JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(nuevoUsuario)
	if err != nil {
		// Si ocurre un error al codificar el JSON, lanzo un código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error codificando a JSON el usuario actualizado", http.StatusInternalServerError)
		return
	}
}

// Funcion para el handler DELETE usuario
func eliminarUsuario(w http.ResponseWriter, r *http.Request, id_usuario int32) {
	// Crea el contexto necesario para las operaciones sqlc
	ctx := r.Context()

	// Elimino el partido
	err := queries.DeleteUsuario(ctx, id_usuario)
	if err != nil {
		// Si ocurre un error al eliminar el usuario, lanzo código 500 y finalizo la ejecucion del handler
		http.Error(w, "Error eliminando partido", http.StatusInternalServerError)
		return
	}

	// Lanzo el código 204 como respuesta exitosa de la operación
	w.WriteHeader(http.StatusNoContent)
}

// Funciones para handlers de Estadisticas
// Funcion para el handler POST Estadistica

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

func validarEstadisticasJugador(estadisticas EstadisticaJugadorRequest) error {
	if estadisticas.Goles == nil || *estadisticas.Goles < 0 {
		return errors.New("El partido no tiene el dato obligatorio: Goles")
	}
	if estadisticas.Asistencias == nil || *estadisticas.Asistencias < 0 {
		return errors.New("El partido no tiene el dato obligatorio: Asistencias")
	}
	if estadisticas.PasesCompletados.Valid {
		// Controlo solo para los valores No Nulo
		return validarFormatoPorcentaje(estadisticas.PasesCompletados, "pases completados")
	}
	if estadisticas.DuelosGanados.Valid {
		// Controlo solo para los valores No Nulo
		return validarFormatoPorcentaje(estadisticas.DuelosGanados, "duelos ganados")
	}

	// Si pasa la validación
	return nil
}

func validarEstadisticasArquero(estadisticas EstadisticaArqueroRequest) error {
	if estadisticas.GolesRecibidos == nil || *estadisticas.GolesRecibidos < 0 {
		return errors.New("El partido no tiene el dato obligatorio: Goles Recibidos")
	}
	if estadisticas.AtajadasClave == nil || *estadisticas.AtajadasClave < 0 {
		return errors.New("El partido no tiene el dato obligatorio: Atajadas Clave")
	}
	if estadisticas.SaquesCompletados.Valid {
		// Controlo solo para los valores No Nulo
		return validarFormatoPorcentaje(estadisticas.SaquesCompletados, "saques completados")
	}

	// Si pasa la validación
	return nil
}

func validarFormatoPorcentaje(dato sql.NullString, atributo string) error {

	// Reemplazo la coma por el punto, en caso de que el formato no fuera el adecuado
	s := strings.ReplaceAll(dato.String, ",", ".")

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
	switch r.Method {
	case "POST":
		//aca deberia estar la funcion que crea una estadistica de jugador
	case "UPDATE":
		//aca deberia estar la funcion que actualiza la estadistica de jugador
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func EstadisticasArqueroHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		//aca deberia estar la funcion que crea una estadistica de jugador
	case "UPDATE":
		//aca deberia estar la funcion que actualiza la estadistica de jugador
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

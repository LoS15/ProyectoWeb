-- name: InsertEstadisticaJugador :one
INSERT INTO Estadistica_Jugador(id_usuario, id_partido, goles, asistencias, pases_completados, duelos_ganados) VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING *;

-- name: UpdateEstadisticaJugador :one
UPDATE Estadistica_Jugador SET goles = $3, asistencias= $4, pases_completados= $5, duelos_ganados= $6
WHERE id_usuario = $1 AND id_partido = $2
RETURNING *;

-- name: GetEstadisticaJugador :one
SELECT * FROM Estadistica_Jugador WHERE id_usuario = $1 AND id_partido = $2;

-- name: DeleteEstadisticaJugador :exec
DELETE FROM Estadistica_Jugador WHERE id_usuario = $1 AND id_partido = $2;
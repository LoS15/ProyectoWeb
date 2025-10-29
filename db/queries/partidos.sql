-- name: InsertPartido :one
INSERT INTO partido(id_usuario, fecha, cancha, puntuacion) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePartido :one
UPDATE partido SET fecha = $3, cancha=$4, puntuacion = $5
WHERE id_usuario = $1 AND id_partido = $2
RETURNING *;

-- name: ListPartidosPorUsuario :many
SELECT * FROM partido WHERE id_usuario = $1;

-- name: DeletePartido :exec
DELETE FROM partido WHERE id_usuario = $1 AND id_partido = $2;

-- name: GetPartidoPorUsuario :one
SELECT * FROM partido WHERE id_usuario = $1 AND id_partido = $2;

-- name: GetAllPartido :many
SELECT * FROM partido;
-- name: InsertPartido :one
INSERT INTO partido(id_usuario, fecha, cancha, puntuacion) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePartido :exec
UPDATE partido SET puntuacion = $2
WHERE id_usuario = $1;

-- name: ListPartidos :many
SELECT * FROM partido WHERE id_usuario = $1;

-- name: DeletePartido :exec
DELETE FROM partido WHERE id_usuario = $1 AND id_partido = $2;

-- name: GetPartido :one
SELECT * FROM partido WHERE id_usuario = $1 AND id_partido = $2;

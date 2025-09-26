-- name: CreateUser :one
INSERT INTO usuario(nombre, apellido, pais) VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUser :exec
UPDATE usuario SET nombre=$2, apellido=$3, pais=$4 WHERE id_usuario = $1;
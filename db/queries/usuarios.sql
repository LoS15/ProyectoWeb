-- name: CreateUser :one
INSERT INTO usuario(nombre, apellido, pais) VALUES ($1, $2, $3)
RETURNING id_usuario, nombre, apellido, pais;

-- name: UpdateUser :one
UPDATE usuario SET nombre=$2, apellido=$3, pais=$4 WHERE id_usuario = $1
RETURNING *;

-- name: GetAllUser :many
SELECT * FROM usuario;

-- name: GetUsuario :one
SELECT * FROM usuario WHERE id_usuario = $1;

-- name: DeleteUsuario :exec
DELETE FROM usuario WHERE id_usuario = $1;
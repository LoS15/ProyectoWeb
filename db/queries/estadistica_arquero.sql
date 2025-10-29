-- name: InsertEstadisticaArquero :one
INSERT INTO Estadistica_Arquero(id_usuario, id_partido, goles_recibidos, atajadas_clave, saques_completados) VALUES ($1, $2, $3, $4, $5)
    RETURNING *;

-- name: UpdateEstadisticaArquero :one
UPDATE Estadistica_Arquero SET goles_recibidos = $3, atajadas_clave= $4, saques_completados= $5
WHERE id_usuario = $1 AND id_partido = $2
RETURNING *;

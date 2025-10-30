#insertar user1
POST http://localhost:8080/usuarios
Content-Type: application/json
{
    "nombre": "Juan",
    "apellido": "Doe",
    "pais": "Argentina"
}
HTTP 201
[Asserts]
jsonpath "$.nombre" == "Juan"
jsonpath "$.apellido" == "Doe"
jsonpath "$.pais" == "Argentina"
[Captures]
id_usuario1: jsonpath "$.id_usuario"

#crear partido1
POST http://localhost:8080/partidos
Content-Type: application/json
{
    "id_usuario": {{id_usuario1}},
    "fecha": "2025-10-28T00:00:00Z",
    "cancha": "Nahuel",
    "puntuacion": 9
}
HTTP 201
[Asserts]
jsonpath "$.fecha" == "2025-10-28T00:00:00Z"
jsonpath "$.cancha" == "Nahuel"
jsonpath "$.puntuacion" == 9
[Captures]
id_usuario_partido1: jsonpath "$.id_usuario"
id_partido1: jsonpath "$.id_partido"

#crear estadisticas_arquero para partido1
POST http://localhost:8080/estadisticas-arquero
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}},
    "goles_recibidos": 1,
    "atajadas_clave": 3,
    "saques_completados": "0.80"
}
HTTP 201
[Asserts]
jsonpath "$.goles_recibidos" == 1
jsonpath "$.atajadas_clave" == 3
jsonpath "$.saques_completados" == "0.80"

#obtiene las estadisticas de manera generica
GET http://localhost:8080/estadisticas/{{id_usuario_partido1}}/{{id_partido1}}
HTTP 200
[Asserts]
jsonpath "$.estadistica_arquero.id_usuario" == {{id_usuario_partido1}}
jsonpath "$.estadistica_arquero.id_partido" == {{id_partido1}}
jsonpath "$.tipo_estadistica" == "arquero"
jsonpath "$.estadistica_arquero.goles_recibidos" == 1
jsonpath "$.estadistica_arquero.atajadas_clave" == 3
jsonpath "$.estadistica_arquero.saques_completados" == "0.80"

#eliminar partido1
DELETE http://localhost:8080/partidos/{{id_usuario_partido1}}/{{id_partido1}}
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}}
}
HTTP 204

#eliminar user1
DELETE http://localhost:8080/usuarios/{{id_usuario1}}
HTTP 204

#verifico que la lista de users este vacia
GET http://localhost:8080/usuarios
HTTP 200
[Asserts]
jsonpath "$[*]" count == 0
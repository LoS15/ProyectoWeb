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

#crear estadisticas_jugador para partido1
POST http://localhost:8080/estadisticas-jugador
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}},
    "goles": 1,
    "asistencias": 2,
    "pases_completados": "0.50",
    "duelos_ganados": "0.60"
}
HTTP 201
[Asserts]
jsonpath "$.goles" == 1
jsonpath "$.asistencias" == 2
jsonpath "$.pases_completados" == "0.50"
jsonpath "$.duelos_ganados" == "0.60"

#modificar estadisticas_jugador
PUT http://localhost:8080/estadisticas-jugador
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}},
    "goles": 2,
    "asistencias": 1,
    "pases_completados": "0.70",
    "duelos_ganados": "0.10"
}
HTTP 201
[Asserts]
jsonpath "$.goles" == 2
jsonpath "$.asistencias" == 1
jsonpath "$.pases_completados" == "0.70"
jsonpath "$.duelos_ganados" == "0.10"

#obtener estadisticas_jugador
GET http://localhost:8080/estadisticas-jugador/{{id_usuario_partido1}}/{{id_partido1}}
HTTP 200
[Asserts]
jsonpath "$.goles" == 2
jsonpath "$.asistencias" == 1
jsonpath "$.pases_completados" == "0.70"
jsonpath "$.duelos_ganados" == "0.10"

#eliminar estadisticas_jugador
DELETE http://localhost:8080/estadisticas-jugador/{{id_usuario_partido1}}/{{id_partido1}}
HTTP 204

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
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
    "fecha": "2025-10-28"
    "cancha": "Nahuel"
    "puntuacion": 9
}
HTTP 201
[Asserts]
jsonpath "$.fecha" == "2025-10-28"
jsonpath "$.cancha" == "Nahuel"
jsonpath "$.puntuacion" == 9
[Captures]
id_usuario_partido1: jsonpath "$.id_usuario"
id_partido1: jsonpath "$.id_partido"

#crear partido2
POST http://localhost:8080/partidos
Content-Type: application/json
{
    "id_usuario": {{id_usuario1}},
    "fecha": "2025-10-29",
    "cancha": "Pico",
    "puntuacion": 7
}
HTTP 201
[Asserts]
jsonpath "$.fecha" == "2025-10-29"
jsonpath "$.cancha" == "Pico"
jsonpath "$.puntuacion" == 7
[Captures]
id_usuario_partido2: jsonpath "$.id_usuario"
id_partido2: jsonpath "$.id_partido"

#crear estadisticas_arquero para partido1
POST http://localhost:8080/estadisticas-arquero
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}},
    "goles_recibidos": 1,
    "atajadas_clave": 3,
    "saques_compleados": 0.8
}
HTTP 200
[Asserts]
jsonpath "$.goles_recibidos" == 1
jsonpath "$.atajadas_clave" == 3
jsonpath "$.saques_completados" == 0.8

#crear estadisticas_jugador para partido2
POST http://localhost:8080/estadisticas-jugador
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido2},
    "id_partido": {{id_partido2}},
    "goles": 2,
    "asistencias": 1,
    "pases_completados": 0.9,
    "duelos_ganados": 0.5
}
HTTP 200
[Asserts]
jsonpath "$.goles" == 1
jsonpath "$.asistencias" == 1
jsonpath "$.pases_completados" == 0.9
jsonpath "$.duelos_ganados" == 0.5

#actualizo estadisticas-arquero de partido1
PUT http://localhost:8080/estadisticas-arquero
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}},
    "goles_recibidos": 2,
    "atajadas_clave": 5,
    "saques_compleados": 0.7
}
HTTP 200
[Asserts]
jsonpath "$.goles_recibidos" == 2
jsonpath "$.atajadas_clave" == 5
jsonpath "$.saques_completados" == 0.7

#actualizar estadisticas_jugador para partido2
PUT http://localhost:8080/estadisticas-jugador
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido2},
    "id_partido": {{id_partido2}},
    "goles": 1,
    "asistencias": 1,
    "pases_completados": 0.4,
    "duelos_ganados": 0.4
}
HTTP 200
[Asserts]
jsonpath "$.goles" == 1
jsonpath "$.asistencias" == 1
jsonpath "$.pases_completados" == 0.4
jsonpath "$.duelos_ganados" == 0.4

#las estadisticas no se pueden eliminar por si mismas, pero podemos hacer que se eliminen por medio de eliminar los partidos (cascade en la database)

#eliminar partido1
DELETE http://localhost:8080/partidos
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}},
}
HTTP 204

#eliminar partido2
DELETE http://localhost:8080/partidos
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido2}},
    "id_partido": {{id_partido2}},
}
HTTP 204

#listar nuevamente partidos (deberia haber 0)
GET http://localhost:8080/partidos
HTTP 200
[Asserts]
jsonpath "$.length" == 0

#eliminar user1
DELETE http://localhost:8080/usuarios/{{id_usuario1}}
HTTP 204

#verifico que la lista de users este vacia
GET http://localhost:8080/usuarios
HTTP 200
[Asserts]
jsonpath "$.length" == 0
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

#crear partido2
POST http://localhost:8080/partidos
Content-Type: application/json
{
    "id_usuario": {{id_usuario1}},
    "fecha": "2025-10-29T00:00:00Z",
    "cancha": "Pico",
    "puntuacion": 7
}
HTTP 201
[Asserts]
jsonpath "$.fecha" == "2025-10-29T00:00:00Z"
jsonpath "$.cancha" == "Pico"
jsonpath "$.puntuacion" == 7
[Captures]
id_usuario_partido2: jsonpath "$.id_usuario"
id_partido2: jsonpath "$.id_partido"

#actualizar partido1
PUT http://localhost:8080/partidos
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}},
    "cancha": "Nahuel_Nuevo",
    "fecha": "2020-10-29T00:00:00Z",
    "puntuacion": 7
}
HTTP 200
[Asserts]
jsonpath "$.fecha" == "2020-10-29T00:00:00Z"
jsonpath "$.cancha" == "Nahuel_Nuevo"
jsonpath "$.puntuacion" == 7

#getPartido1 (luego de haberlo actualizado)
GET http://localhost:8080/partidos/{{id_usuario_partido1}}/{{id_partido1}}
HTTP 200
[Asserts]
jsonpath "$.fecha" == "2020-10-29T00:00:00Z"
jsonpath "$.cancha" == "Nahuel_Nuevo"
jsonpath "$.puntuacion" == 7

#listar partidos de todos los users
GET http://localhost:8080/partidos
HTTP 200
[Asserts]
jsonpath "$[*]" count == 2

#eliminar partido1
DELETE http://localhost:8080/partidos/{{id_usuario_partido1}}/{{id_partido1}}
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido1}},
    "id_partido": {{id_partido1}}
}
HTTP 204

#eliminar partido2
DELETE http://localhost:8080/partidos/{{id_usuario_partido2}}/{{id_partido2}}
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido2}},
    "id_partido": {{id_partido2}}
}
HTTP 204

#listar nuevamente (deberia haber 0)
GET http://localhost:8080/partidos
HTTP 200
[Asserts]
jsonpath "$[*]" count == 0

#eliminar user1
DELETE http://localhost:8080/usuarios/{{id_usuario1}}
HTTP 204

#verifico que la lista de users este vacia
GET http://localhost:8080/usuarios
HTTP 200
[Asserts]
jsonpath "$[*]" count == 0
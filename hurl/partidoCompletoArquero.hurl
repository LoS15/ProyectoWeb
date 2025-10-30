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

#crear partido completo
POST http://localhost:8080/crearPartido
Content-Type: application/json
{
    "id_usuario": {{id_usuario1}},
    "fecha": "2018-10-29",
    "cancha": "Nahuel",
    "puntuacion": 9,
    "tipo_estadistica":"arquero",
    "estadistica_jugador": null,
    "estadistica_arquero" : {
            "goles_recibidos": 2,
            "atajadas_clave": 1,
            "saques_completados": 0.9
        }
}
HTTP 201
[Asserts]
jsonpath "$.id_usuario" == {{id_usuario1}},
jsonpath "$.fecha" == "2018-10-29T00:00:00Z",
jsonpath "$.cancha" == "Nahuel",
jsonpath "$.puntuacion" == 9,
jsonpath "$.tipo_estadistica":"arquero",
jsonpath "$.estadistica_arquero.goles_recibidos" == 2
jsonpath "$.estadistica_arquero.atajadas_clave" == 1
jsonpath "$.estadistica_arquero.saques_completados" == 0.9
[Captures]
id_partido1 : jsonpath "$.id_partido"

#eliminar partido
DELETE http://localhost:8080/partidos/{{id_usuario1}}/{{id_partido1}}
Content-Type: application/json
{
    "id_usuario": {{id_usuario_partido2}},
    "id_partido": {{id_partido2}}
}
HTTP 204

#eliminar user1
DELETE http://localhost:8080/usuarios/{{id_usuario1}}
HTTP 204
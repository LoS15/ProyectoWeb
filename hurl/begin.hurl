# Insertar user1
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
id_usuario1: jsonpath "$['id_usuario']"

#crear partido completo1
POST http://localhost:8080/crearPartido
Content-Type: application/json
{
    "id_usuario": {{id_usuario1}},
    "fecha": "2018-10-29T00:00:00Z",
    "cancha": "Nahuel",
    "puntuacion": 9,
    "tipo_estadistica":"arquero",
    "estadistica_jugador": null,
    "estadistica_arquero" : {
            "goles_recibidos": 2,
            "atajadas_clave": 1,
            "saques_completados": "0.90"
        }
}
HTTP 201
[Asserts]
jsonpath "$.id_usuario" == {{id_usuario1}}
jsonpath "$.fecha" == "2018-10-29T00:00:00Z"
jsonpath "$.cancha" == "Nahuel"
jsonpath "$.puntuacion" == 9
jsonpath "$.tipo_estadistica" == "arquero"
jsonpath "$.estadistica_arquero.goles_recibidos" == 2
jsonpath "$.estadistica_arquero.atajadas_clave" == 1
jsonpath "$.estadistica_arquero.saques_completados" == "0.90"
[Captures]
id_partido1 : jsonpath "$.id_partido"

#crear partido completo2
POST http://localhost:8080/crearPartido
Content-Type: application/json
{
    "id_usuario": {{id_usuario1}},
    "fecha": "2018-10-29T00:00:00Z",
    "cancha": "Nahuel",
    "puntuacion": 9,
    "tipo_estadistica": "jugador",
    "estadistica_jugador" : {
        "goles": 2,
        "asistencias": 1,
        "pases_completados": "0.90",
        "duelos_ganados": "0.50"
    },
    "estadistica_arquero": null
}
HTTP 201
[Asserts]
jsonpath "$.id_usuario" == {{id_usuario1}}
jsonpath "$.fecha" == "2018-10-29T00:00:00Z"
jsonpath "$.cancha" == "Nahuel"
jsonpath "$.puntuacion" == 9
jsonpath "$.tipo_estadistica" == "jugador"
jsonpath "$.estadistica_jugador.goles" == 2
jsonpath "$.estadistica_jugador.asistencias" == 1
jsonpath "$.estadistica_jugador.pases_completados" == "0.90"
jsonpath "$.estadistica_jugador.duelos_ganados" == "0.50"
[Captures]
id_partido1 : jsonpath "$.id_partido"
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
    "tipo_estadistica":"jugador",
    "estadistica_jugador" : {
        "goles": 2,
        "asistencias": 1,
        "pases_completados": 0.9,
        "duelos_ganados": 0.5
    },
    "estadistica_arquero": null
}
HTTP 201
[Asserts]

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

# Insertar user2
POST http://localhost:8080/usuarios
Content-Type: application/json
{
    "nombre": "Jane",
    "apellido": "Doe",
    "pais": "United States"
}
HTTP 201
[Asserts]
jsonpath "$.nombre" == "Jane"
jsonpath "$.apellido" == "Doe"
jsonpath "$.pais" == "United States"
[Captures]
id_usuario2: jsonpath "$['id_usuario']"

# Obtener user1
GET http://localhost:8080/usuarios/{{id_usuario1}}
HTTP 200
[Asserts]
jsonpath "$.nombre" == "Juan"
jsonpath "$.apellido" == "Doe"
jsonpath "$.pais" == "Argentina"

# Actualizar user2
PUT http://localhost:8080/usuarios
Content-Type: application/json

{
    "id_usuario": {{id_usuario2}},
    "nombre": "Juan_Nuevo",
    "apellido": "Doe_Nuevo",
    "pais": "Argentina_Nuevo"
}

HTTP 200
[Asserts]
jsonpath "$.nombre" == "Juan_Nuevo"
jsonpath "$.apellido" == "Doe_Nuevo"
jsonpath "$.pais" == "Argentina_Nuevo"

# Listar todos los users
GET http://localhost:8080/usuarios
HTTP 200
[Asserts]
jsonpath "$[*]" count == 2

# Eliminar user1
DELETE http://localhost:8080/usuarios/{{id_usuario1}}
HTTP 204

# Eliminar user2
DELETE http://localhost:8080/usuarios/{{id_usuario2}}
HTTP 204

# Verifico que la lista esté vacía
GET http://localhost:8080/usuarios
HTTP 200
[Asserts]
jsonpath "$[*]" count == 0
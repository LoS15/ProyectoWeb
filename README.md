Conformacion de grupo:  
  Juan Martin Lopez Vannoni  
  Iñaki Uranga Vega

Practico 1 --> abarca el directorio static y los archivos libres (excepto el sqlc.yaml)  
Practico 2 --> abarca el directorio db y de los archivos libres, solo el archivo sqlc.yaml

Bitacora de Futbol (nombre preliminar)  
Es una pagina web dedicada a que cada futbolista pueda registrar su rendimiento en cada partido que juega, donde hay estadisticas tanto para jugadores de campo como para arqueros.
La idea de utilizacion de la pagina web es la siguiente:
  1. El usuario se registra en la pagina web.
  2. El usuario registra un partido, con sus estadisticas como arquero y/o jugador de campo.
  3. El usuario luego puede consultar todos sus partidos o tambien un subconjunto de ellos a partir de diferentes criterios (entre fechas, por rendimiento, por cantidad de goles, etc.).

La implementacion de estos pasos aun no se ha hecho, ya que pertenecen a etapas posteriores de este proyecto.

Estructura del proyecto:
  - Subdirectorio db: contiene el subdirectorio queries (incluye queries sobre las tablas usuario y partido) y el subdirectorio schema (define el esquema de la base de datos)
  - Subdirectorio static: aqui se encuentra el archivo index.html y su respectivo style.css, que se presentan en el root (/) de la pagina web.
  - Directorio princpial: main.go y handlers.go, se utilizan para correr el servidor. El archivo docker-compose.yml se utiliza, por el momento, solo para levantar una imagen de postgresql en docker. El archivo sqlc.yaml se utiliza para ejecutar el       sqlc generate.

Como ejecutar el servidor web del practico 1:
  1. En el directorio /ProyectoWeb, abrir CLI y ejecutar comando: go run main.go handlers.go
  2. En su navegador de confianza: http://localhost:8080/

Como generar los archivos del sqlc (ya estan generados, pero si se quisieran regenerar) del practico 2:
  1. En el directorio /ProyectoWeb, abrir CLI y ejecutar comando: sqlc generate

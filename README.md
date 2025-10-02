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

Como ejecutar el servidor web del practico 1:
  1. En la CLI: go run main.go handlers.go
  2. En su navegador de confianza: http://localhost:8080/

Como generar los archivos del sqlc (ya estan generados, pero si se quisieran regenerar)
  1. En la CLI: sqlc generate

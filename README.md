# Instrucciones para iniciar la aplicacion (valido para TP3 y TP4):

El script *runApp.sh* corre los 8 test de hurl que hicimos y deja corriendo ambos containers para probar la aplicacion.

## Como acceder al frontend:
Para acceder a la pagina web, simplemente puede introducir esta URL en su navegador de preferencia: *http://localhost:8080/*

## Como finalizar la ejecucion de la aplicacion:
Una vez que quiera bajar los containers, debe introducir el siguiente comando en la CLI: *docker compose -f docker-compose.app.yml down -v* (la opcion -v es para eliminar la persistencia interna dentro de los containers, para que cada ejecucion del docker compose sea independiente de las demas).

# Aclaraciones:

## Sobre herramientas:
La herramienta que utilizamos para crear los containers a traves de docker compose es la nueva version de docker compose que viene incluida con docker CLI, y el comando a utilizar es "docker compose", en vez de "docker-compose" que es la herramienta independiente hecha en Python.

## Sobre desarrollo del proyecto:

### Correcciones sobre TPs anteriores:

- En la tabla Partido añadimos id_usuario como PK. Además, lo hicimos de tipo int en lugar de serial.

- En ambas tablas de Estadísticas añadimos el atributo id_usuario (por ahora ser PK de partido junto con id_partido), lo hicimos PK Y FK de estas tablas.

- Los atributos que podián ser no null, ahora son obligatorios.

### Sobre TP3 y TP4:

Hay desarrollos que se hicieron a modo de prueba para el Hurl, debido a que se pidieron ciertas operaciones para todas entidades pero que luego en el contexto del uso de la aplicación, no tendrían razon de ser (sea porque el flujo de uso de la misma no lo requiere o por que existe una autenticación y autorización que limita ciertos usos y hace innecesarias ciertas operaciones). Estos estan marcados con comentarios del estilo a:
// [nombre y mínima explicacion de que es el desarrollo] *(*solo para pruebas*)*
Por ejemplo:
// Funciones para handlers de EstadisticasJugador *(*solo para pruebas*)*


#### Políticas de validación
Se usaron tipos auxiliares con variacion en el tipo de cada campo según la política tomada y según lo que necesitamos:

- Para todos los campos int32 debido a que son obligatorios, se utilizo el tipo auxiliar *int32 para validar que dicho campo no sea nil (el campo es inexistente o el valor del campo es null) o que tenga un valor válido (0 > caso de ID's y 0 >= para el resto de casos). Claramente, si el campo existe y el valor es "" no se puede decodificar y dara error, lo cual esta bien. No se dejo el int normal sabiendo que si el campo es inexiste o el valor del campo es null, por defecto se lo transforma en un campo con valor 0, porque como mencione antes hay campos que si permiten el valor 0 como válido (y habría casos que no sabríamos si es por un error o porque verdaderamente se buscaba ese valor).

- Para todos los campos time.Time debido a que son obligatorios, se utilizo el tipo auxiliar *time.Time para validar que dicho campo no sea nil (el campo es inexistente o el valor del campo es null) o que tenga un valor válido (distinto de 0001-01-01 00:00:00 +0000 UTC). Desde el front, se concatena una hora default para que funcione, otra opcion podria haber sido trabajar con fechas con hora en la base de datos, pero nos parecio irrelevante la hora y como Go no tiene otro tipo para fechas que no sea time.Time que si utiliza hora, no nos quedo de otra.

- Para todos los campos string obligatorios, se utilizo el mismo tipo string sabiendo que si el campo es inexistente o el valor del campo es null, por defecto se lo transforma en un campo con valor "", entonces simplemente se chequea que el campo no sea "" (por los dos causantes anteriormente mencionados o porque simplemente el campo existe y se lo completo con valor vacío).

- Para los strings que son representacion de flotantes, estan controlados los errores de flotantes escritos con coma, en cuyo caso se reemplaza la coma por un punto. Por otro lado, tambien se controlan los errores de flotantes que se encuentran entre el 0 y el 1 (representando porcentajes) y que fueran escritos con 1 solo digito decimal, para ellos se le agrega un cero al final, para poder insertarlos correctamente en la base de datos.


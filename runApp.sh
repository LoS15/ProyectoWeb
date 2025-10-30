#!/bin/bash

# --- Configuración ---
# 1. Salir inmediatamente si cualquier comando falla (set -e)
# 2. Tratar las variables no seteadas como un error (set -u)
# 3. El 'pipefail' hace que un pipeline falle si cualquier comando en él falla
set -euo pipefail

# --- Limpieza en caso de Error ---
# Esta función se ejecutará si el script falla en cualquier punto (gracias a 'trap')
cleanup_on_failure() {
    echo ""
    echo "--- ❌ ¡Las pruebas fallaron! ---"
    echo "--- Bajando todos los servicios... ---"
    docker compose -f docker-compose.app.yml down -v
    exit 1
}

# 'trap' le dice a bash: "Si recibís una señal de error (ERR), ejecutá la función 'cleanup_on_failure'"
trap cleanup_on_failure ERR

# --- 1. Construir y Levantar Servicios ---
echo "--- Construyendo y levantando servicios (API + DB) ---"
# Usamos '--build' para asegurarnos de que la imagen de la API esté actualizada
# Usamos '-d' (detached) para que corra en segundo plano
docker compose -f docker-compose.app.yml up -d --build

# --- 2. Esperar que la DB esté lista ---
echo "--- Esperando a que los servicios inicien (5 segundos)... ---"
sleep 5

# --- 3. Ejecutar Pruebas ---
echo "--- Corriendo tests de Hurl ---"
# Ejecutamos los tests contra los servicios que acabamos de levantar
for f in hurl/*.hurl; do
    echo "--- Ejecutando $f ---"
    hurl --test "$f"
done

# --- 4. Éxito ---
# Si llegamos aquí, las pruebas pasaron.
# Desactivamos el 'trap' para que no se ejecute la limpieza al salir
trap - ERR

echo ""
echo "--- ¡Pruebas finalizadas con éxito! ---"
echo "--- La aplicación (API y DB) sigue corriendo. ---"
echo ""
echo "Estado actual de los servicios:"
docker compose -f docker-compose.app.yml ps
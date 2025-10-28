const selectorJugador = document.getElementById('tipo_jugador')
const selectorArquero = document.getElementById('tipo_arquero')
const formularioJugador = document.getElementById('estadisticas-jugador-campo')
const formularioArquero = document.getElementById('estadisticas-arquero')

selectorJugador.addEventListener('change', (event) => {
    if (selectorJugador.checked){
        formularioJugador.style.display = 'block'
        formularioArquero.style.display = 'none'
    }
})

selectorArquero.addEventListener('change', (event) => {
    if (selectorArquero.checked){
        formularioJugador.style.display = 'none'
        formularioArquero.style.display = 'block'
    }
})

const formularioPartido = document.getElementById('formulario-partido')
const feedbackFormulario = document.getElementById('feedback-formulario')

formularioPartido.addEventListener('submit', async(event) => {
    event.preventDefault();  // previene que se envie por default
    feedback.textContent = 'Enviando formulario...';
    //creo un objeto plano para rellenar con los campos pertinentes
    const objetoFormulario = new FormData(formularioPartido);
    const data = Object.fromEntries(formularioPartido.entries());
    //construyo el json inicial, sin tener aun estadisticas de jugador o arquero
    const json_partido = {
        id_usuario: parseInt(data.id_usuario),
        fecha: data.fecha,
        cancha: data.cancha,
        tipo_estadistica: data.tipo_estadistica,
        estadistica_jugador: null,
        estadistica_arquero: null,
    }
    //ahora lleno el objeto con lo que corresponde
    if (json_partido.tipo_estadistica === 'tipo_jugador') {
        json_partido.estadistica_jugador = {
            goles:parseInt(data.goles),
            asistencias:parseInt(data.asistencias),
            pases_completados:data.pases_completados,
            duelos_ganados:data.duelos_ganados
        }
    } else {
        json_partido.estadistica_arquero = {
            goles_recibidos: parseInt(data.goles_recibidos),
            atajadas_clave: parseInt(data.atajadas_clave),
            saques_completados: data.saques_completados
        }
    }
    //ahora envio el JSON con fetch
    try {
        const response = await fetch('/crearPartido', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json', //declaramos que lo enviado es un JSON
            },
            body: JSON.stringify(json_partido) //convertimos el string que teniamos antes en el JSON a enviar
        });
        if (response.ok) {
            const result = await response.text();
            feedbackFormulario.textContent = 'El formulario ha sido enviado.';
            feedbackFormulario.style.color = 'green';
            formularioPartido.reset();
            cargarPartidos()
        } else {
            const error = await response.text();
            feedbackFormulario.textContent = `Error de red: ${error}.`;
            feedbackFormulario.style.color = 'red';
        }
    } catch (error) {
        feedbackFormulario.textContent = `Error de red: ${error.message}.`;
        feedbackFormulario.style.color = 'red';
    }
})

//
document.addEventListener('DOMContentLoaded', () => {
    cargarPartidos();
})

async function cargarPartidos() {
    const divLista = document.getElementById('div-partidos');
    try {
        const response = await fetch('/partidos'); //GET para partidos
        if (!response.ok) {
            throw new Error(`Error HTTP! Status: ${response.status}`);
        }
        const partidos = await response.json();
        //ahora muestro partidos
        divLista.innerHTML = '';
        if (partidos.length === 0) { //check que haya partidos existentes
            divLista.innerHTML = 'No hay partidos registrados.';
            return;
        }
        partidos.forEach(partido => {
            const elementoPartido = document.createElement('div');
            elementoPartido.className = 'elemento-partido';
            //creo elemento de informacion
            const infoPartido = document.createElement('p');
            infoPartido.textContent = 'ID_Usuario' + partido.id_usuario + '.ID_Partido: ' + partido.id_partido + '. Fecha: ' + partido.fecha + '. Cancha: ' + partido.cancha + '. Puntuacion: ' + partido.puntuacion + '.';
            //creo boton de estadisticas
            const botonEstadisticas = document.createElement('button');
            botonEstadisticas.textContent = 'Ver estadisticas';
            botonEstadisticas.className = 'boton-estadisticas';
            //guardo los datos que uso para buscar las estadisticas
            botonEstadisticas.dataset.idPartido = partido.id_partido;
            botonEstadisticas.dataset.idUsuario = partido.id_usuario;
            //creo boton de eliminar partido
            const botonEliminar = document.createElement('button');
            botonEliminar.textContent = 'Eliminar partido';
            botonEliminar.className = 'boton-eliminar-partido';
            //guardo los datos que uso para eliminar partido
            botonEliminar.dataset.idPartido = partido.id_partido;
            botonEliminar.dataset.idUsuario = partido.id_usuario;
            //crea el div donde estaran las estadisticas
            const divEstadistica = document.createElement('div');
            divEstadistica.className = 'div-estadisticas';
            //armamos el elementoPartido
            elementoPartido.appendChild(infoPartido);
            elementoPartido.appendChild(botonEstadisticas);
            elementoPartido.appendChild(divEstadistica);
            elementoPartido.appendChild(botonEliminar);
            divLista.appendChild(elementoPartido);
        })
    } catch (error) {
        divLista.textContent = `Error: ${error.message}.`;
    }
}

const listaPartidos = document.getElementById('div-partidos');
listaPartidos.addEventListener('click', async(event)=> {
    if(event.target.classList.contains('boton-estadisticas')){
        const boton = event.target; //con esto puedo acceder a las variables del elemento donde ocurrio el evento
        const partidoId = boton.dataset.idPartido;
        const usuarioId = boton.dataset.idUsuario;
        //ahora busco el div que contiene a este boton, para poder modificarlo luego
        const divEstadistica = boton.parentElement.querySelector('.div-estadisticas');
        if (!divEstadistica) return; // esto tengo q fijarme
        try {
            const response = await fetch(`/estadisticas?id_usuario=${usuarioId}&id_partido=${partidoId}`,
                method='GET',
                );
            if (!response.ok) {
                throw new Error(`No se encontraron estadisticas.`); //no puede suceder en teoria, pero bueno (REVISAR)
            }
            const estadisticas = await response.json();
            divEstadistica.innerHTML = ''; //limpio el html (lo inicio mas bien)
            const informacionEstadistica = document.createElement('div'); //debo hacer esta division para agregar el boton luego y este div como hijos
            if (estadisticas.tipo === 'jugador'){ //son estadisticas de jugador
                informacionEstadistica.innerHTML = ' <strong>Estadisticas Jugador</strong>' +
                    '<p>Goles: ${estadisticas.goles}</p>' +
                    '<p>Asistencias: ${estadisticas.asistencias}</p>'+
                    '<p>Pases completados: ${estadisticas.pases_completados}</p>'+
                    '<p>Duelos ganados: ${estadisticas.duelos_ganados}</p>';
            } else {
                informacionEstadistica.innerHTML = ' <strong>Estadisticas Jugador</strong>' +
                    '<p>Goles recibidos: ${estadisticas.goles_recibidos}</p>' +
                    '<p>Atajadas clave: ${estadisticas.atajadas_clave}</p>'+
                    '<p>Saques completados: ${estadisticas.saques_completados}</p>';
            }
            const botonModificarEstadistica = document.createElement('button');
            botonModificarEstadistica.textContent = 'Modificar estadistica';
            botonModificarEstadistica.className = 'boton-modificar-estadistica';
            botonModificarEstadistica.dataset.json = JSON.stringify(estadisticas.datos);
            divEstadistica.appendChild(informacionEstadistica);
            divEstadistica.appendChild(botonModificarEstadistica);
        } catch (error) {
            divEstadistica.textContent = `Error: ${error.message}.`;
        }
    } else if (event.target.classList.contains('boton-eliminar-partido')) {
        const boton = event.target; //con esto puedo acceder a las variables del elemento donde ocurrio el evento
        const partidoId = event.dataset.idPartido;
        const usuarioId = event.dataset.idUsuario;
        if (!confirm('Esta seguro de querer eliminar este partido?')){
            return;
        }
        try {
            const response = await fetch(`/partidos?id_usuario=${usuarioId}&id_partido=${partidoId}`,
                method='DELETE',
            );
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error); //cuando puede suceder esto?
            }
            console.log("Partido eliminado. Refrescando lista.");
            cargarPartidos();
        } catch (error) {
            alert(`Error al eliminar partido: ${error.message}`);
        }
    } else if (event.target.classList.contains('boton-modificar-estadistica')) {
        const boton = event.target;
        const tipoEstadistica = boton.dataset.tipo;
        const estadisticas = JSON.parse(boton.dataset.json);
        const divEstadistica = boton.closest('.div-estadisticas'); //obtiene ancestro mas cercano de tipo div-estadisticas
        // crear formulario aqui (debo hacerlo para inye)
        let formularioModifEstadistica = '';
        if (tipo === 'jugador') {
            formularioModifEstadistica= `
                <form class="form-modificar-estadistica">
                    <input type="hidden" name="tipo" value="jugador">
                    <input type="hidden" name="id_partido" value="${stats.id_partido}">
                    <input type="hidden" name="id_usuario" value="${stats.id_usuario}">
                    <label>Goles:</label>
                    <input type="number" name="goles" value="${stats.goles}">
                    <label>Asistencias:</label>
                    <input type="number" name="asistencias" value="${stats.asistencias}">
                    <button type="submit">Guardar Cambios</button>
                    <button type="button" class="btn-cancelar-mod">Cancelar</button>
                </form>
            `;
        } else {
            formularioModifEstadistica= `
                <form class="form-modificar-estadistica">
                    <input type="hidden" name="tipo" value="arquero">
                    <input type="hidden" name="id_partido" value="${stats.id_partido}">
                    <input type="hidden" name="id_usuario" value="${stats.id_usuario}">
                    <label>Goles Recibidos:</label>
                    <input type="number" name="goles_recibidos" value="${stats.goles_recibidos}">
                    <button type="submit">Guardar Cambios</button>
                    <button type="button" class="btn-cancelar-mod">Cancelar</button>
                </form>
            `;
        }
        divEstadistica.innerHTML = formularioModifEstadistica;
    } else if (event.target.classList.contains('form-modificar-estadistica')) {
        event.preventDefault();
        const formModificarEstadisticas = event.target;
        const dataFormulario = new FormData(formModificarEstadisticas);
        const data = Object.fromEntries(dataFormulario.entries());
        //seguir maniana el evento de modificar estadistica. Tambien falta boton cancelar y boton modificar partido.
    }
})
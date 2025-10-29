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
    feedbackFormulario.textContent = 'Enviando formulario...';
    //creo un objeto plano para rellenar con los campos pertinentes
    const data = Object.fromEntries(formularioPartido.entries());
    //construyo el json inicial, sin tener aun estadisticas de jugador o arquero
    const json_partido = {
        id_usuario: parseInt(data.id_usuario),
        fecha: data.fecha + "T00:00:00Z",
        cancha: data.cancha,
        tipo_estadistica: data.tipo_estadistica,
        puntuacion : data.puntuacion,
        estadistica_jugador: null,
        estadistica_arquero: null,
    }
    //ahora lleno el objeto con lo que corresponde
    if (json_partido.tipo_estadistica === 'jugador') {
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
            botonEstadisticas.type = 'button';
            //guardo los datos que uso para buscar las estadisticas
            botonEstadisticas.dataset.idPartido = partido.id_partido;
            botonEstadisticas.dataset.idUsuario = partido.id_usuario;
            //creo boton de eliminar partido
            const botonEliminar = document.createElement('button');
            botonEliminar.textContent = 'Eliminar partido';
            botonEliminar.className = 'boton-eliminar-partido';
            botonEliminar.type = 'button';
            //guardo los datos que uso para eliminar partido
            botonEliminar.dataset.idPartido = partido.id_partido;
            botonEliminar.dataset.idUsuario = partido.id_usuario;
            //creo boton de modificar partido
            const botonModificarPartido = document.createElement('button');
            botonModificarPartido.textContent = 'Modificar partido';
            botonModificarPartido.className = 'boton-modificar-partido';
            botonModificarPartido.type = 'button'; //lo hago para asegurarme que no haga submit
            //guardo los datos necesarios para modificar el usuario
            botonModificarPartido.dataset.json = JSON.stringify(partido); //le doy todo para q lo use al momento de modificar inplace
            //crea el div donde estaran las estadisticas
            const divEstadistica = document.createElement('div');
            divEstadistica.className = 'div-estadisticas';
            //armamos el elementoPartido
            elementoPartido.appendChild(infoPartido);
            elementoPartido.appendChild(botonEstadisticas);
            elementoPartido.appendChild(divEstadistica);
            elementoPartido.appendChild(botonEliminar);
            elementoPartido.appendChild(botonModificarPartido);
            divLista.appendChild(elementoPartido);
        })
    } catch (error) {
        divLista.textContent = `Error: ${error.message}.`;
    }
}

const listaPartidos = document.getElementById('div-partidos');
listaPartidos.addEventListener('click', async(event)=> {
    const target = event.target;
    if(target.classList.contains('boton-estadisticas')){
        mostrarEstadisticasDePartido(target);
    } else if (target.classList.contains('boton-eliminar-partido')) {
        eliminarPartido(target);
    } else if (target.classList.contains('boton-modificar-estadistica')) {
        mostrarFormularioModificacionEstadisticas(target);
        //seguir maniana el evento de modificar estadistica. Tambien falta boton cancelar y boton modificar partido.
    } else if (target.classList.contains('boton-modificar-partido')) { //ACCION SOBRE BOTON MODIFICAR PARTIDO
        mostrarFormularioModificacionPartido(target);
    } else if (target.classList.contains('boton-cancelar-mod')) {
        const div_estadistica = target.closest('.div-estadisticas');
        if (div_estadistica) { //si es estadistica, solo borro el contenido del div_estadistica, menos pesado
            div_estadistica.innerText = '';
            console.log("Cancelando modificacion sobre estadisticas de partido.")
            return;
        }
        const elemento_partido = target.closest('.elemento-partido');
        if (elemento_partido) {
            console.log("Cancelando modificacion sobre partido. Refrescando lista de partidos...");
            cargarPartidos();
        }
    }
});

listaPartidos.addEventListener('submit', async(event)=> { //listener solo sobre submit para formularios
    event.preventDefault();
    if (event.target.classList.contains('formulario-modificar-estadisticas')) { //ACCION SOBRE RECEPCION DE FORM DE MODIFICAR ESTADISTICAS
        modificarEstadisticasPartido(event.target)
    } else if (event.target.classList.contains('formulario-modificar-partido')) {
        modificarPartido(event.target);
    }
})


async function mostrarEstadisticasDePartido(boton){
    const partidoId = boton.dataset.idPartido;
    const usuarioId = boton.dataset.idUsuario;
    //ahora busco el div que contiene a este boton, para poder modificarlo luego
    const divEstadistica = boton.parentElement.querySelector('.div-estadisticas');
    if (!divEstadistica) return; // esto tengo q fijarme
    try {
        const response = await fetch(`/estadisticas?id_usuario=${usuarioId}&id_partido=${partidoId}`, {
                method: 'GET'
            }
        );
        if (!response.ok) {
            throw new Error(`No se encontraron estadisticas.`); //no puede suceder en teoria, pero bueno (REVISAR)
        }
        const estadisticas = await response.json();
        divEstadistica.innerHTML = ''; //limpio el html (lo inicio mas bien)
        const informacionEstadistica = document.createElement('div'); //debo hacer esta division para agregar el boton luego y este div como hijos
        if (estadisticas.tipo === 'jugador'){ //son estadisticas de jugador
            informacionEstadistica.innerHTML = ` <strong>Estadisticas Jugador</strong>' +
                '<p>Goles: ${estadisticas.goles}</p>' +
                '<p>Asistencias: ${estadisticas.asistencias}</p>'+
                '<p>Pases completados: ${estadisticas.pases_completados}</p>'+
                '<p>Duelos ganados: ${estadisticas.duelos_ganados}</p>`;
        } else {
            informacionEstadistica.innerHTML = ` <strong>Estadisticas Jugador</strong>' +
                '<p>Goles recibidos: ${estadisticas.goles_recibidos}</p>' +
                '<p>Atajadas clave: ${estadisticas.atajadas_clave}</p>'+
                '<p>Saques completados: ${estadisticas.saques_completados}</p>`;
        }
        const botonModificarEstadistica = document.createElement('button');
        botonModificarEstadistica.textContent = 'Modificar estadistica';
        botonModificarEstadistica.className = 'boton-modificar-estadistica';
        botonModificarEstadistica.dataset.json = JSON.stringify(estadisticas.data);
        divEstadistica.appendChild(informacionEstadistica);
        divEstadistica.appendChild(botonModificarEstadistica);
    } catch (error) {
        divEstadistica.textContent = `Error: ${error.message}.`;
    }
}

async function eliminarPartido(boton){
    const partidoId = boton.dataset.idPartido;
    const usuarioId = boton.dataset.idUsuario;
    if (!confirm('Esta seguro de querer eliminar este partido?')){
        return;
    }
    try {
        const response = await fetch(`/partidos?id_usuario=${usuarioId}&id_partido=${partidoId}`, {
            method: 'DELETE'
        });
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error); //cuando puede suceder esto?
        }
        console.log("Partido eliminado. Refrescando lista.");
        cargarPartidos();
    } catch (error) {
        alert(`Error al eliminar partido: ${error.message}`);
    }
}

async function mostrarFormularioModificacionEstadisticas(boton){
    const estadisticas = JSON.parse(boton.dataset.json);
    const tipoEstadistica = estadisticas.tipo;
    const divEstadistica = boton.closest('.div-estadisticas'); //obtiene ancestro mas cercano de tipo div-estadisticas
    // crear formulario aqui (a traves de template definidos)
    let template;
    if (tipoEstadistica === 'jugador'){
        template = document.getElementById('template-formulario-jugador');
    } else {
        template = document.getElementById('template-formulario-arquero');
    }
    //ahora creo una copia del template que puede ingresarse al DOM de la pagina
    const templateDom = template.content.cloneNode(true);
    //relleno los placeholders con los valores que le corresponde, eso hace que pueda editar los valores in place, e intentar luego enviar la modificacion a la BD
    templateDom.querySelector('[name="id_usuario"]').value = estadisticas.id_usuario;
    templateDom.querySelector('[name="id_partido"]').value = estadisticas.id_partido;
    if (tipoEstadistica === 'jugador'){
        templateDom.querySelector('[name="goles"]').value = estadisticas.goles;
        templateDom.querySelector('[name="asistencias"]').value = estadisticas.asistencias;
        templateDom.querySelector('[name="pases-completados"]').value = estadisticas.pases_completados;
        templateDom.querySelector('[name="duelos-ganados"]').value = estadisticas.duelos_ganados;
    } else {
        templateDom.querySelector('[name="goles-recibidos"]').value = estadisticas.goles_recibidos;
        templateDom.querySelector('[name="atajadas-clave"]').value = estadisticas.atajadas_clave;
        templateDom.querySelector('[name="saques-completados"]').value = estadisticas.saques_completados;
    }
    //ahora asigno el formulario al innerHTML,
    divEstadistica.innerHTML = '';
    divEstadistica.appendChild(templateDom);
}

async function modificarEstadisticasPartido(formulario){
    const dataFormulario = new FormData(formulario); //obtiene los pares clave-valor del form
    const data = Object.fromEntries(dataFormulario.entries()); //convierte esos clave-valor a un objeto javascript plano
    const tipoEstadistica = data.tipo;
    let jsonEstadistica;
    //guardamos en una variable el div que contiene a las estadisticas, para poder comunicar como procede la modificacion
    const divEstadistica = formulario.parentElement;
    divEstadistica.innerHTML = 'Guardando cambios...';
    //diferenciamos entre jugador y arquero para armar el JSON
    if (tipoEstadistica === 'jugador'){
        jsonEstadistica = {
            id_usuario: parseInt(data.id_usuario),
            id_partido: parseInt(data.id_partido),
            goles: parseInt(data.goles),
            asistencias: parseInt(data.asistencias),
            pases_completados: data.pases_completados,
            duelos_ganados: data.duelos_ganados
        }
    } else {
        jsonEstadistica = {
            id_usuario: parseInt(data.id_usuario),
            id_partido: parseInt(data.id_partido),
            goles_recibidos: parseInt(data.goles_recibidos),
            atajadas_clave: parseInt(data.atajadas_clave),
            saques_completados: data.saques_completados,
        }
    }
    //ahora que el JSON esta listo, llamo a la funcion para enviar a guardar los cambios
    try {
        const response = await fetch('/estadisticas', { //revisar el endpoint mas adelante
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(jsonEstadistica),
        });
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }
        console.log("Modificacion exitosa. Recargando lista"); //consultar si debemos refrescar toda la lista o solo el elemento modificado
        cargarPartidos()
    } catch (error) {
        divEstadistica.innerHTML = `Error al intentar modificar las estadisticas: ${error.message}`;
    }
}

async function mostrarFormularioModificacionPartido(boton) {
    const info_partido = JSON.parse(boton.dataset.json);
    const divPartido = boton.closest('.elemento-partido'); //obtiene ancestro mas cercano de tipo div-partido, es decir, la carta que contiene a ese partido
    // crear formulario aqui (a traves de template definidos)
    const template = document.getElementById("template-formulario-partido");
    //ahora creo una copia del template que puede ingresarse al DOM de la pagina
    const templateDom = template.content.cloneNode(true);
    //relleno los placeholders con los valores que le corresponde, eso hace que pueda editar los valores in place, e intentar luego enviar la modificacion a la BD
    templateDom.querySelector('[name="id_usuario"]').value = info_partido.id_usuario;
    templateDom.querySelector('[name="id_partido"]').value = info_partido.id_partido;
    templateDom.querySelector('[name="fecha_partido"]').value = new Date(info_partido.fecha).toISOString().split('T')[0]; //hay que convertirla en yy/mm/dd
    templateDom.querySelector('[name="cancha"]').value = info_partido.cancha;
    templateDom.querySelector('[name="puntuacion"]').value = info_partido.puntuacion;
    //ahora asigno el formulario al innerHTML,
    divPartido.innerHTML = '';
    divPartido.appendChild(templateDom);
}

async function modificarPartido(formulario){
    const dataFormulario = new FormData(formulario); //obtiene los pares clave-valor del form
    const data = Object.fromEntries(dataFormulario.entries()); //convierte esos clave-valor a un objeto javascript plano
    const jsonPartido = {
        id_usuario: parseInt(data.id_usuario),
        id_partido: parseInt(data.id_partido),
        fecha_partido: data.fecha_partido,
        cancha: data.cancha,
        puntuacion: parseInt(data.puntuacion),
    };
    //guardamos en una variable el div que contiene al partido, para poder comunicar como procede la modificacion
    const divPartido = formulario.parentElement;
    divPartido.innerHTML = 'Guardando cambios...';
    //ahora que el JSON esta listo, llamo a la funcion para enviar a guardar los cambios
    try {
        const response = await fetch(`/partidos/${data.id_usuario}/${data.id_partido}`, { //revisar el endpoint mas adelante
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(jsonPartido),
        });
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }
        console.log("Modificacion exitosa. Recargando lista"); //consultar si debemos refrescar toda la lista o solo el elemento modificado
        cargarPartidos()
    } catch (error) {
        divPartido.innerHTML = `Error al intentar modificar el partido: ${error.message}`;
    }
}
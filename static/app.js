document.addEventListener('DOMContentLoaded', () => { //CREO UN EVENT LISTENER PARA CARGAR LOS PARTIDOS APENAS SE ABRE LA PAGINA
    let flagPartidosCargadosInicialmente = false;

    /*SECCION DE NAVEGACION*/

    //secciones
    const seccionUsuario = document.getElementById('container-formulario-usuario');
    const seccionPartidos = document.getElementById('container-seccion-partidos');

    //botones de navegacion del header
    const botonUsuario = document.getElementById('boton-usuario');
    const botonPartidos = document.getElementById('boton-partidos');

    //escucha de eventos sobre botones de navegacion del header
    botonUsuario.addEventListener('click', (e) => {
        mostrarVista('seccion-usuario');
    })
    botonPartidos.addEventListener('click', (e) => {
        mostrarVista('seccion-partidos');
    })

    //funcion de vistas
    function mostrarVista(vistaDada){
        if (vistaDada === 'seccion-partidos'){
            seccionPartidos.style.display = 'block';
            seccionUsuario.style.display = 'none';
            if(!flagPartidosCargadosInicialmente){
                flagPartidosCargadosInicialmente = true;
                cargarPartidos();
            }
        } else {
            seccionUsuario.style.display = 'block';
            seccionPartidos.style.display = 'none';
        }
    }

    /*SECCION USUARIO*/

    //formulario de usuario (secciones)
    const formularioUsuario = document.getElementById('formulario-usuario');
    const feedbackFormularioUsuario = document.getElementById('feedback-formulario-usuario');

    //evento sobre submit de creacion de usuario
    formularioUsuario.addEventListener('submit', async(event) => {
        event.preventDefault();  // previene que se envie por default
        feedbackFormularioUsuario.textContent = 'Enviando formulario...';
        //obtengo la info de las entries del formulario, para armar el JSON de la request
        const data = Object.fromEntries(new FormData(formularioUsuario));
        //armo el JSON
        const jsonUsuario = {
            "nombre": data.nombre,
            "apellido": data.apellido,
            "pais": data.pais,
        }
        try {
            const response = await fetch('/usuarios', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json', //declaramos que lo enviado es un JSON
                },
                body: JSON.stringify(jsonUsuario) //convertimos el string que teniamos antes en el JSON a enviar
            });
            if (!response.ok) {
                const error = response.statusText;
                feedbackFormularioUsuario.textContent = `Ha ocurrido un error al intentar registrar el usuario. Error: ${error}`;
                feedbackFormularioUsuario.style.color = 'red';
            } else {
                feedbackFormularioUsuario.textContent = 'El formulario ha sido enviado correctamente.';
                feedbackFormularioUsuario.style.color = 'green';
            }
        } catch (error) {
            feedbackFormularioUsuario.textContent = `Error de red: ${error.message}.`;
            feedbackFormularioUsuario.style.color = 'red';
        }
    })

    /*SECCION DE PARTIDOS*/

    //selectores de jugador o arquero, para el formulario de partido completo
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

    //formulario y evento de creacion de partido completo
    const formularioPartido = document.getElementById('formulario-partido')
    const feedbackFormularioPartido = document.getElementById('feedback-formulario-partido')

    formularioPartido.addEventListener('submit', async(event) => { //FORMULARIO PARA ENVIAR PARTIDO_COMPLETO
        event.preventDefault();  // previene que se envie por default
        feedbackFormularioPartido.textContent = 'Enviando formulario...';
        //obtengo la info de las entries del formulario, para armar el JSON de la request
        const data = Object.fromEntries(new FormData(formularioPartido));//construyo el json inicial, sin tener aun estadisticas de jugador o arquero
        const json_partido = {
            id_usuario: parseInt(data.id_usuario),
            fecha: data.fecha_partido + "T00:00:00Z",
            cancha: data.cancha,
            puntuacion : parseInt(data.puntuacion),
            tipo_estadistica: data.tipo_estadistica,
            estadistica_jugador: null,
            estadistica_arquero: null,
        }
        //ahora lleno el objeto con lo que corresponde
        if (json_partido.tipo_estadistica === 'jugador') {
            json_partido.estadistica_jugador = {
                goles:parseInt(data.goles),
                asistencias:parseInt(data.asistencias),
                pases_completados: data.pases_completados,
                duelos_ganados: data.duelos_ganados
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
                feedbackFormularioPartido.textContent = 'El formulario ha sido enviado correctamente.';
                feedbackFormularioPartido.style.color = 'green';
                formularioPartido.reset(); //reseteo el formulario
                cargarPartidos()
            } else {
                const error = await response.text();
                feedbackFormularioPartido.textContent = `Error de red: ${error}.`;
                feedbackFormularioPartido.style.color = 'red';
            }
        } catch (error) {
            feedbackFormularioPartido.textContent = `Error de red: ${error.message}.`;
            feedbackFormularioPartido.style.color = 'red';
        }
    })

    //seccion de lista de partidos, y evento de listener sobre mostrar estadisticas, eliminar partidos, eliminar estadisticas o mostrar formularios para modificar estadisticas o partidos
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

    //listener de modificaciones sobre estadisticas o partidos
    listaPartidos.addEventListener('submit', async(event)=> { //listener solo sobre submit para formularios
        event.preventDefault();
        if (event.target.classList.contains('formulario-modificar-estadisticas')) { //ACCION SOBRE RECEPCION DE FORM DE MODIFICAR ESTADISTICAS
            modificarEstadisticasPartido(event.target)
        } else if (event.target.classList.contains('formulario-modificar-partido')) {
            modificarPartido(event.target);
        }
    })

})

async function cargarPartidos() {
    const divLista = document.getElementById('div-partidos');
    try {
        const response = await fetch('/partidos', {method:'GET'}); //GET para partidos
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
            infoPartido.textContent = `ID Usuario: ${partido.id_usuario}. ID Partido: ${partido.id_partido}. Fecha: ${partido.fecha}. Cancha: ${partido.cancha}. Puntuación: ${partido.puntuacion}.`;
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

async function mostrarEstadisticasDePartido(boton){
    const partidoId = boton.dataset.idPartido;
    const usuarioId = boton.dataset.idUsuario;
    //ahora busco el div que contiene a este boton, para poder modificarlo luego
    const divEstadistica = boton.parentElement.querySelector('.div-estadisticas');
    if (!divEstadistica) return; // esto tengo q fijarme
    try {
        const response = await fetch(`/estadisticas/${usuarioId}/${partidoId}`, {
                method: 'GET'
            }
        );
        if (!response.ok) {
            throw new Error(`No se encontraron estadisticas.`); //no puede suceder en teoria, pero bueno (REVISAR)
        }
        const estadisticas = await response.json();
        divEstadistica.innerHTML = ''; //limpio el html (lo inicio mas bien)
        const informacionEstadistica = document.createElement('div'); //debo hacer esta division para agregar el boton luego y este div como hijos
        if (estadisticas.tipo_estadistica === 'jugador'){ //son estadisticas de jugador
            informacionEstadistica.innerHTML = ` <strong>Estadisticas Jugador</strong>
                <p>Goles: ${estadisticas.estadistica_jugador.goles}</p> 
                <p>Asistencias: ${estadisticas.estadistica_jugador.asistencias}</p>
                <p>Pases completados: ${estadisticas.estadistica_jugador.pases_completados}</p>
                <p>Duelos ganados: ${estadisticas.estadistica_jugador.duelos_ganados}</p>`;
        } else {
            informacionEstadistica.innerHTML = `<strong>Estadisticas Arquero</strong>
                <p>Goles recibidos: ${estadisticas.estadistica_arquero.goles_recibidos}</p>
                <p>Atajadas clave: ${estadisticas.estadistica_arquero.atajadas_clave}</p>
                <p>Saques completados: ${estadisticas.estadistica_arquero.saques_completados}</p>`;
        }
        const botonModificarEstadistica = document.createElement('button');
        botonModificarEstadistica.textContent = 'Modificar estadistica';
        botonModificarEstadistica.className = 'boton-modificar-estadistica';
        botonModificarEstadistica.dataset.json = JSON.stringify(estadisticas);
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
        const response = await fetch(`/partidos/${usuarioId}/${partidoId}`, {
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
    const tipoEstadistica = estadisticas.tipo_estadistica;
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
    if (tipoEstadistica === 'jugador'){
        templateDom.querySelector('[name="id_usuario"]').value = estadisticas.estadistica_jugador.id_usuario;
        templateDom.querySelector('[name="id_partido"]').value = estadisticas.estadistica_jugador.id_partido;
        templateDom.querySelector('[name="goles"]').value = estadisticas.estadistica_jugador.goles;
        templateDom.querySelector('[name="asistencias"]').value = estadisticas.estadistica_jugador.asistencias;
        templateDom.querySelector('[name="pases_completados"]').value = estadisticas.estadistica_jugador.pases_completados;
        templateDom.querySelector('[name="duelos_ganados"]').value = estadisticas.estadistica_jugador.duelos_ganados;
    } else {
        templateDom.querySelector('[name="id_usuario"]').value = estadisticas.estadistica_arquero.id_usuario;
        templateDom.querySelector('[name="id_partido"]').value = estadisticas.estadistica_arquero.id_partido;
        templateDom.querySelector('[name="goles_recibidos"]').value = estadisticas.estadistica_arquero.goles_recibidos;
        templateDom.querySelector('[name="atajadas_clave"]').value = estadisticas.estadistica_arquero.atajadas_clave;
        templateDom.querySelector('[name="saques_completados"]').value = estadisticas.estadistica_arquero.saques_completados;
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
    const jsonArquero = data.arquero;
    let call_path;
    if (tipoEstadistica === 'jugador'){
        jsonEstadistica = {
            id_usuario: parseInt(data.id_usuario),
            id_partido: parseInt(data.id_partido),
            goles: parseInt(data.goles),
            asistencias: parseInt(data.asistencias),
            pases_completados: data.pases_completados,
            duelos_ganados: data.duelos_ganados
        }
        call_path = `/estadisticas-jugador`;
    } else {
        jsonEstadistica = {
            id_usuario: parseInt(data.id_usuario),
            id_partido: parseInt(data.id_partido),
            goles_recibidos: parseInt(data.goles_recibidos),
            atajadas_clave: parseInt(data.atajadas_clave),
            saques_completados: data.saques_completados,
        }
        call_path = `/estadisticas-arquero`;
    }
    //ahora que el JSON esta listo, llamo a la funcion para enviar a guardar los cambios
    try {
        const response = await fetch(call_path, {
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(jsonEstadistica),
        });
        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        } else {
            console.log("Modificacion exitosa. Recargando lista"); //consultar si debemos refrescar toda la lista o solo el elemento modificado
            cargarPartidos()
        }
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
    templateDom.querySelector('[name="fecha_partido"]').value = new Date(info_partido.fecha).toISOString().split('T')[0];//hay que convertirla en yy/mm/dd
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
        fecha: data.fecha_partido + "T00:00:00Z",
        cancha: data.cancha,
        puntuacion: parseInt(data.puntuacion),
    };
    //guardamos en una variable el div que contiene al partido, para poder comunicar como procede la modificacion
    const divPartido = formulario.parentElement;
    divPartido.innerHTML = 'Guardando cambios...';
    //ahora que el JSON esta listo, llamo a la funcion para enviar a guardar los cambios
    try {
        const response = await fetch(`/partidos`, { //revisar el endpoint mas adelante
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
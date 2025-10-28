CREATE TABLE Estadistica_Arquero (
    id_usuario int NOT NULL,
    id_partido int  NOT NULL,
    goles_recibidos int  NOT NULL,
    atajadas_clave int  NOT NULL,
    saques_completados decimal(3,2)  NULL,
    CONSTRAINT PK_ESTADISTICA_ARQUERO PRIMARY KEY (id_usuario, id_partido)
);

CREATE TABLE Estadistica_Jugador (
    id_usuario int NOT NULL,
    id_partido int  NOT NULL,
    goles int  NOT NULL,
    asistencias int  NOT NULL,
    pases_completados decimal(3,2)  NULL,
    duelos_ganados decimal(3,2)  NULL,
    CONSTRAINT PK_ESTADISTICA_JUGADOR PRIMARY KEY (id_usuario, id_partido)
);

CREATE TABLE Partido (
    id_partido serial NOT NULL,
    id_usuario int NOT NULL,
    fecha date  NOT NULL,
    cancha varchar(35)  NOT NULL,
    puntuacion int  NOT NULL,
    CONSTRAINT PK_PARTIDO PRIMARY KEY (id_usuario, id_partido)
);

CREATE TABLE Usuario (
    id_usuario serial  NOT NULL,
    nombre varchar(50)  NOT NULL,
    apellido varchar(50)  NOT NULL,
    pais varchar(35)  NOT NULL,
    CONSTRAINT PK_USUARIO PRIMARY KEY (id_usuario)
);

ALTER TABLE Estadistica_Arquero ADD CONSTRAINT FK_ESTADISTICA_ARQUERO_PARTIDO
    FOREIGN KEY (id_usuario, id_partido)
    REFERENCES Partido (id_usuario, id_partido)
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

ALTER TABLE Estadistica_Jugador ADD CONSTRAINT FK_ESTADISTICA_JUGADOR_PARTIDO
    FOREIGN KEY (id_usuario, id_partido)
    REFERENCES Partido (id_usuario, id_partido)
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

ALTER TABLE Partido ADD CONSTRAINT FK_PARTIDO_USUARIO
    FOREIGN KEY (id_usuario)
    REFERENCES Usuario (id_usuario)  
    NOT DEFERRABLE 
    INITIALLY IMMEDIATE
;

--Todos los atributos con decimal(3, 2) es porque fueron pensados como porcentajes, valores entre 0,00 y 1,00.

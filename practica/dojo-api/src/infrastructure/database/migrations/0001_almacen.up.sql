-- Compatible con PostgreSQL 10 o superior.
CREATE TABLE IF NOT EXISTS almacen (
    id         serial       PRIMARY KEY,
    clave      varchar(20)  NOT NULL UNIQUE,
    nombre     varchar(120) NOT NULL,
    activo     boolean      NOT NULL DEFAULT true,
    created_at timestamptz  NOT NULL DEFAULT now(),
    updated_at timestamptz  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id              SERIAL,
    nickname        TEXT UNIQUE NOT NULL,
    first_name      TEXT NOT NULL,
    last_name       TEXT NOT NULL,
    country         TEXT NOT NULL,
    password        TEXT NOT NULL,
    email           TEXT UNIQUE NOT NULL,
    disabled        BOOL DEFAULT 'f',
    version         INT DEFAULT 1,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY(id)
)
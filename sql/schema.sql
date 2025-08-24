CREATE TABLE IF NOT EXISTS config_defaults (
    id INTEGER PRIMARY KEY,
    name VARCHAR NOT NULL,
    file_name VARCHAR NOT NULL,
    path VARCHAR NOT NULL,
    program VARCHAR NOT NULL,
    content VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS registered_programs (
    id INTEGER PRIMARY KEY,
    program VARCHAR NOT NULL UNIQUE,
    file_name VARCHAR NOT NULL
);

-- name: InsertConfigDefaults :exec
INSERT INTO config_defaults
    (name, file_name, path, program, content)
VALUES
    (?, ?, ?, ?, ?);

-- name: GetConfigDefaultByName :one
SELECT content, name, file_name, path FROM config_defaults WHERE name = ?;

-- name: GetAllConfigs :many
SELECT content, name, path FROM config_defaults;

-- name: DeleteConfigByName :exec
DELETE FROM config_defaults WHERE name = ?;

-- name: RegisterNewConfigFile :exec
INSERT INTO registered_programs
    (program, file_name)
VALUES
    (?, ?);

-- name: InsertDefaultRegisteredPrograms :exec
INSERT OR IGNORE INTO registered_programs
    (program, file_name)
VALUES
    ('sqlc', 'sqlc.yaml');

-- name: GetRegisteredPrograms :many
SELECT program, file_name FROM registered_programs;

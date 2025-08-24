-- name: InsertConfigDefaults :exec
INSERT INTO config_defaults
    (name, file_name, program, content)
VALUES
    (?, ?, ?, ?);

-- name: GetConfigDefaultByName :one
SELECT content, name, file_name FROM config_defaults WHERE name = ?;

-- name: GetAllConfigs :many
SELECT content, name FROM config_defaults;


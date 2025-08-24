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

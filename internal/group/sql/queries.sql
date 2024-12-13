-- name: GetAllReports :many
SELECT 
    report_text,
    reporter_id,
    reported_user_id,
    origin_server_id
FROM reports;

-- name: AddServer :exec
INSERT INTO servers (
    server_id
) VALUES (?);

-- name: AddUser :exec
INSERT OR IGNORE INTO users (
    user_id
) VALUES (?);

-- name: CreateReport :one
INSERT INTO reports (
    report_text,
    reporter_id, 
    reported_user_id,
    origin_server_id
) values (?, ?, ?, ?)
RETURNING report_id;


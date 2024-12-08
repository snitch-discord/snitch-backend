-- name: CreateUser :exec
INSERT INTO users (user_id) VALUES (?);

-- name: CreateServer :exec
INSERT INTO servers (server_id) VALUES (?);

-- name: CreateReport :one
INSERT INTO reports (
    report_text,
    reporter_id,
    reported_user_id,
    origin_server_id
) VALUES (?, ?, ?, ?) RETURNING *;

-- name: GetReport :one
SELECT * FROM reports
WHERE report_id = ? LIMIT 1;

-- name: UpdateReport :exec
UPDATE reports
SET report_text = ?
WHERE report_id = ?;

-- name: DeleteReport :exec
DELETE FROM reports
WHERE report_id = ?;

-- name: ListUserReports :many
SELECT r.*, 
    reporter.user_id as reporter_user_id,
    reported.user_id as reported_user_id,
    s.server_id as origin_server_id
FROM reports r
JOIN users reporter ON r.reporter_id = reporter.user_id
JOIN users reported ON r.reported_user_id = reported.user_id
JOIN servers s ON r.origin_server_id = s.server_id
WHERE r.reported_user_id = ?
ORDER BY r.report_id DESC;

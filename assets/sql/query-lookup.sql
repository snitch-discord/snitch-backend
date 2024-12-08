-- name: GetGroup :one
SELECT * FROM groups
WHERE group_id = ? LIMIT 1;

-- name: ListGroups :many
SELECT * FROM groups
ORDER BY group_id;

-- name: CreateGroup :one
INSERT INTO groups (
    group_id,
    group_name
) VALUES (?, ?) RETURNING *;

-- name: UpdateGroup :exec
UPDATE groups
SET group_name = ?
WHERE group_id = ?;

-- name: DeleteGroup :exec
DELETE FROM groups
WHERE group_id = ?;

-- name: GetServerGroup :one
SELECT g.*
FROM groups g
JOIN servers s ON s.group_id = g.group_id
WHERE s.server_id = ?;

-- name: CreateServer :one
INSERT INTO servers (
    server_id,
    output_channel,
    group_id,
    permission_level
) VALUES (?, ?, ?, ?) RETURNING *;

-- name: DeleteServer :exec
DELETE FROM servers
WHERE server_id = ? AND group_id = ?;

-- name: GetServerCount :one
SELECT COUNT(*) FROM servers
WHERE group_id = ?;

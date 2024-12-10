-- name: GetAllReports :many
SELECT 
    report_text,
    reporter_id,
    reported_user_id,
    origin_server_id
FROM reports;

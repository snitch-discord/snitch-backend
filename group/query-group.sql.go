// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query-group.sql

package group

import (
	"context"
)

const createReport = `-- name: CreateReport :one
INSERT INTO reports (
    report_text,
    reporter_id,
    reported_user_id,
    origin_server_id
) VALUES (?, ?, ?, ?) RETURNING report_id, report_text, reporter_id, reported_user_id, origin_server_id
`

type CreateReportParams struct {
	ReportText     string `json:"report_text"`
	ReporterID     int64  `json:"reporter_id"`
	ReportedUserID int64  `json:"reported_user_id"`
	OriginServerID int64  `json:"origin_server_id"`
}

func (q *Queries) CreateReport(ctx context.Context, arg CreateReportParams) (Report, error) {
	row := q.queryRow(ctx, q.createReportStmt, createReport,
		arg.ReportText,
		arg.ReporterID,
		arg.ReportedUserID,
		arg.OriginServerID,
	)
	var i Report
	err := row.Scan(
		&i.ReportID,
		&i.ReportText,
		&i.ReporterID,
		&i.ReportedUserID,
		&i.OriginServerID,
	)
	return i, err
}

const createServer = `-- name: CreateServer :exec
INSERT INTO servers (server_id) VALUES (?)
`

func (q *Queries) CreateServer(ctx context.Context, serverID int64) error {
	_, err := q.exec(ctx, q.createServerStmt, createServer, serverID)
	return err
}

const createUser = `-- name: CreateUser :exec
INSERT INTO users (user_id) VALUES (?)
`

func (q *Queries) CreateUser(ctx context.Context, userID int64) error {
	_, err := q.exec(ctx, q.createUserStmt, createUser, userID)
	return err
}

const deleteReport = `-- name: DeleteReport :exec
DELETE FROM reports
WHERE report_id = ?
`

func (q *Queries) DeleteReport(ctx context.Context, reportID int64) error {
	_, err := q.exec(ctx, q.deleteReportStmt, deleteReport, reportID)
	return err
}

const getReport = `-- name: GetReport :one
SELECT report_id, report_text, reporter_id, reported_user_id, origin_server_id FROM reports
WHERE report_id = ? LIMIT 1
`

func (q *Queries) GetReport(ctx context.Context, reportID int64) (Report, error) {
	row := q.queryRow(ctx, q.getReportStmt, getReport, reportID)
	var i Report
	err := row.Scan(
		&i.ReportID,
		&i.ReportText,
		&i.ReporterID,
		&i.ReportedUserID,
		&i.OriginServerID,
	)
	return i, err
}

const listUserReports = `-- name: ListUserReports :many
SELECT r.report_id, r.report_text, r.reporter_id, r.reported_user_id, r.origin_server_id, 
    reporter.user_id as reporter_user_id,
    reported.user_id as reported_user_id,
    s.server_id as origin_server_id
FROM reports r
JOIN users reporter ON r.reporter_id = reporter.user_id
JOIN users reported ON r.reported_user_id = reported.user_id
JOIN servers s ON r.origin_server_id = s.server_id
WHERE r.reported_user_id = ?
ORDER BY r.report_id DESC
`

type ListUserReportsRow struct {
	ReportID         int64  `json:"report_id"`
	ReportText       string `json:"report_text"`
	ReporterID       int64  `json:"reporter_id"`
	ReportedUserID   int64  `json:"reported_user_id"`
	OriginServerID   int64  `json:"origin_server_id"`
	ReporterUserID   int64  `json:"reporter_user_id"`
	ReportedUserID_2 int64  `json:"reported_user_id_2"`
	OriginServerID_2 int64  `json:"origin_server_id_2"`
}

func (q *Queries) ListUserReports(ctx context.Context, reportedUserID int64) ([]ListUserReportsRow, error) {
	rows, err := q.query(ctx, q.listUserReportsStmt, listUserReports, reportedUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListUserReportsRow
	for rows.Next() {
		var i ListUserReportsRow
		if err := rows.Scan(
			&i.ReportID,
			&i.ReportText,
			&i.ReporterID,
			&i.ReportedUserID,
			&i.OriginServerID,
			&i.ReporterUserID,
			&i.ReportedUserID_2,
			&i.OriginServerID_2,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateReport = `-- name: UpdateReport :exec
UPDATE reports
SET report_text = ?
WHERE report_id = ?
`

type UpdateReportParams struct {
	ReportText string `json:"report_text"`
	ReportID   int64  `json:"report_id"`
}

func (q *Queries) UpdateReport(ctx context.Context, arg UpdateReportParams) error {
	_, err := q.exec(ctx, q.updateReportStmt, updateReport, arg.ReportText, arg.ReportID)
	return err
}

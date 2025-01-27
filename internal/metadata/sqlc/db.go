// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.addServerToGroupStmt, err = db.PrepareContext(ctx, addServerToGroup); err != nil {
		return nil, fmt.Errorf("error preparing query AddServerToGroup: %w", err)
	}
	if q.createGroupTableStmt, err = db.PrepareContext(ctx, createGroupTable); err != nil {
		return nil, fmt.Errorf("error preparing query CreateGroupTable: %w", err)
	}
	if q.createServerTableStmt, err = db.PrepareContext(ctx, createServerTable); err != nil {
		return nil, fmt.Errorf("error preparing query CreateServerTable: %w", err)
	}
	if q.findGroupIDByServerIDStmt, err = db.PrepareContext(ctx, findGroupIDByServerID); err != nil {
		return nil, fmt.Errorf("error preparing query FindGroupIDByServerID: %w", err)
	}
	if q.insertGroupStmt, err = db.PrepareContext(ctx, insertGroup); err != nil {
		return nil, fmt.Errorf("error preparing query InsertGroup: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.addServerToGroupStmt != nil {
		if cerr := q.addServerToGroupStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing addServerToGroupStmt: %w", cerr)
		}
	}
	if q.createGroupTableStmt != nil {
		if cerr := q.createGroupTableStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createGroupTableStmt: %w", cerr)
		}
	}
	if q.createServerTableStmt != nil {
		if cerr := q.createServerTableStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createServerTableStmt: %w", cerr)
		}
	}
	if q.findGroupIDByServerIDStmt != nil {
		if cerr := q.findGroupIDByServerIDStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing findGroupIDByServerIDStmt: %w", cerr)
		}
	}
	if q.insertGroupStmt != nil {
		if cerr := q.insertGroupStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertGroupStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                        DBTX
	tx                        *sql.Tx
	addServerToGroupStmt      *sql.Stmt
	createGroupTableStmt      *sql.Stmt
	createServerTableStmt     *sql.Stmt
	findGroupIDByServerIDStmt *sql.Stmt
	insertGroupStmt           *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                        tx,
		tx:                        tx,
		addServerToGroupStmt:      q.addServerToGroupStmt,
		createGroupTableStmt:      q.createGroupTableStmt,
		createServerTableStmt:     q.createServerTableStmt,
		findGroupIDByServerIDStmt: q.findGroupIDByServerIDStmt,
		insertGroupStmt:           q.insertGroupStmt,
	}
}

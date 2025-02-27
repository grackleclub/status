// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package db

import (
	"context"
	"time"
)

const addStatus = `-- name: AddStatus :exec
INSERT INTO status (
    ts,
    url,
    status_code,
    response_ms
) VALUES (?, ?, ?, ?)
`

type AddStatusParams struct {
	Ts         time.Time
	Url        string
	StatusCode int64
	ResponseMs int64
}

func (q *Queries) AddStatus(ctx context.Context, arg AddStatusParams) error {
	_, err := q.db.ExecContext(ctx, addStatus,
		arg.Ts,
		arg.Url,
		arg.StatusCode,
		arg.ResponseMs,
	)
	return err
}

const stats = `-- name: Stats :many
SELECT ts, url, status_code, response_ms
FROM status
`

func (q *Queries) Stats(ctx context.Context) ([]Status, error) {
	rows, err := q.db.QueryContext(ctx, stats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Status
	for rows.Next() {
		var i Status
		if err := rows.Scan(
			&i.Ts,
			&i.Url,
			&i.StatusCode,
			&i.ResponseMs,
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

const statuses = `-- name: Statuses :many
SELECT ts, url, status_code, response_ms
FROM status
WHERE ts >= ?1
    AND ts < ?2
ORDER BY ts DESC
`

type StatusesParams struct {
	AtAfter time.Time
	Before  time.Time
}

func (q *Queries) Statuses(ctx context.Context, arg StatusesParams) ([]Status, error) {
	rows, err := q.db.QueryContext(ctx, statuses, arg.AtAfter, arg.Before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Status
	for rows.Next() {
		var i Status
		if err := rows.Scan(
			&i.Ts,
			&i.Url,
			&i.StatusCode,
			&i.ResponseMs,
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

const statusesByURL = `-- name: StatusesByURL :many
SELECT ts, url, status_code, response_ms
FROM status
WHERE url = ?1
    AND ts >= ?2
    AND ts < ?3
ORDER BY ts DESC
`

type StatusesByURLParams struct {
	Url     string
	AtAfter time.Time
	Before  time.Time
}

func (q *Queries) StatusesByURL(ctx context.Context, arg StatusesByURLParams) ([]Status, error) {
	rows, err := q.db.QueryContext(ctx, statusesByURL, arg.Url, arg.AtAfter, arg.Before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Status
	for rows.Next() {
		var i Status
		if err := rows.Scan(
			&i.Ts,
			&i.Url,
			&i.StatusCode,
			&i.ResponseMs,
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

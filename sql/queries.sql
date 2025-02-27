-- name: AddStatus :exec
INSERT INTO status (
    ts,
    url,
    status_code,
    response_ms
) VALUES (?, ?, ?, ?);

-- name: Statuses :many
SELECT ts, url, status_code, response_ms
FROM status
WHERE ts >= :atAfter
    AND ts < :before
ORDER BY ts DESC;

-- name: StatusesByURL :many
SELECT ts, url, status_code, response_ms
FROM status
WHERE url = :url
    AND ts >= :atAfter
    AND ts < :before
ORDER BY ts DESC;

-- name: Stats :many
SELECT ts, url, status_code, response_ms
FROM status;

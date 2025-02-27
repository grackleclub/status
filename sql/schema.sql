-- DROP TABLE IF EXISTS status;
CREATE TABLE IF NOT EXISTS status (
    ts TIMESTAMP NOT NULL,
    url TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_ms INTEGER NOT NULL
);

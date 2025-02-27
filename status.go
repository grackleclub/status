package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/grackleclub/status/sql/db"
)

var targets = []string{
	"https://api.grackle.club",
	"https://katlukens.com",
	"https://turkosaur.us",
}

type result struct {
	Url        string // http url
	StatusCode int    // http status code (or 0 if error)
	Rtt        int    // round trip time in milliseconds
	Err        error  // unused
}

// status is a single blocking url check
func status(url string) (int, int, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return 0, int(time.Since(start).Milliseconds()), err
	}
	return resp.StatusCode, int(time.Since(start).Milliseconds()), nil
}

// statuses is a concurrent check of multiple urls, writing results to db
func statuses(ctx context.Context, urls []string) error {
	results := make(chan result)
	for _, url := range urls {
		go func(url string) {
			code, rtt, err := status(url)
			results <- result{url, code, rtt, err}
		}(url)
	}
	for range urls {
		result := <-results
		data := db.AddStatusParams{
			Ts:         time.Now().UTC(),
			Url:        result.Url,
			StatusCode: int64(result.StatusCode),
			ResponseMs: int64(result.Rtt),
		}
		if result.Err != nil {
			slog.Error("fetch failed",
				"url", result.Url,
				"rtt", result.Rtt,
				"err", result.Err,
			)
		}
		err := queries.AddStatus(ctx, data)
		if err != nil {
			return fmt.Errorf("write status for %q: %w", result.Url, err)
		}
		slog.Debug("wrote status", "url", result.Url, "code", result.StatusCode, "rtt", result.Rtt)
	}
	close(results)
	return nil
}

func statusesForever(ctx context.Context, urls []string, interval time.Duration) {
	for {
		err := statuses(ctx, urls)
		if err != nil {
			slog.Error("statuses failed", "err", err)
		}
		time.Sleep(interval)
	}
}

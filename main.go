package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"text/template"
	"time"

	"github.com/grackleclub/status/sql/db"
	_ "modernc.org/sqlite"
)

// lookback defines how many hourly chunks should be included in display
const lookback int = 48

//go:embed static
var static embed.FS

var (
	queries         *db.Queries
	dbStr           = "file:status.db?cache=shared&mode=rw"
	portDefault     = "8888"
	intervalDefault = 5 * time.Minute
)

func init() {
	_, ok := os.LookupEnv("DEBUG")
	if ok {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

func main() {
	ctx := context.Background()
	conn, err := sql.Open("sqlite", dbStr)
	if err != nil {
		panic(fmt.Errorf("open db: %w", err))
	}
	defer conn.Close()
	queries = db.New(conn)

	f, err := os.ReadFile(path.Join("sql", "schema.sql"))
	if err != nil {
		panic(fmt.Errorf("read schema: %w", err))
	}
	_, err = conn.Exec(string(f))
	if err != nil {
		panic(fmt.Errorf("exec migrations: %w", err))
	}

	// continue to check async and write to db
	go statusesForever(ctx, targets, intervalDefault)

	// listen and serve
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = portDefault
	}
	slog.Info("listening", "port", port)

	http.HandleFunc("/", serve)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(fmt.Errorf("listen: %w", err))
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	slog.Debug("request",
		"url", r.URL.Path,
		"method", r.Method,
		"address", r.RemoteAddr,
	)

	p := db.StatusesParams{
		AtAfter: time.Now().UTC().AddDate(0, 0, -7), // 7 days ago
		Before:  time.Now().UTC(),
	}
	statuses, err := queries.Statuses(ctx, p)
	if err != nil {
		slog.Error(
			"statuses failed",
			"err", err,
		)
		http.Error(w, "darn, something went wrong", http.StatusInternalServerError)
		return
	}

	type pings struct {
		Time  time.Time
		Ups   int
		Downs int
	}
	// blocks of time (probably hourly) with lots of pings inside
	// type blocks map[time.Time]pings
	// a map indexed on identifier (probably URL)
	type report map[string][]pings

	type rows struct {
		Time string
		Url  string
		Code int
		Rtt  int
	}
	type page struct {
		Report report
		Rows   []rows
	}
	var pageData page

	var rpts = make(report)
	for _, s := range statuses {
		hour := s.Ts.Truncate(time.Hour)
		p := pings{Time: hour}
		if len(rpts[s.Url]) == 0 {
			rpts[s.Url] = []pings{p}
		} else {
			found := false
			for i := range rpts[s.Url] {
				if rpts[s.Url][i].Time == hour {
					p = rpts[s.Url][i]
					found = true
					break
				}
			}
			if !found {
				rpts[s.Url] = append(rpts[s.Url], p)
			}
		}
		if s.StatusCode == http.StatusOK {
			p.Ups++
		} else {
			p.Downs++
		}
		for i := range rpts[s.Url] {
			if rpts[s.Url][i].Time == hour {
				rpts[s.Url][i] = p
				break
			}
		}
	}

	// Generate time values starting from the current hour and going back an hour each time
	now := time.Now().UTC().Truncate(time.Hour)
	var timesInRange []time.Time
	for i := range lookback {
		timesInRange = append(timesInRange, now.Add(-time.Duration(i)*time.Hour).Truncate(time.Hour))
	}
	slog.Info("generated lookback table", "data", timesInRange)

	// Fill in empty space,
	// chunks are usually an hour
	for url, chunk := range rpts {
		slog.Debug("one", "k", url, "report", chunk)
		for _, interval := range timesInRange {
			var exists bool
			for _, ping := range chunk {
				if ping.Time.Truncate(time.Hour) == interval.Truncate(time.Hour) {
					exists = true
				}
			}
			if !exists {
				slog.Debug("adding empty slot", "url", url, "interval", interval)
				chunk = append(chunk, pings{Time: interval, Ups: 0, Downs: 0})
			}
			rpts[url] = chunk
		}
	}
	pageData.Report = rpts

	// TODO make empty unkown slots for existing services
	// for every services that was up since

	var rawPages []rows
	for _, s := range statuses {

		rawPages = append(rawPages, rows{
			Time: s.Ts.Format(time.RFC3339),
			Url:  s.Url,
			Code: int(s.StatusCode),
			Rtt:  int(s.ResponseMs),
		})
	}
	pageData.Rows = rawPages

	tmpl, err := template.ParseFS(static, path.Join("static", "index.html"))
	if err != nil {
		slog.Error(
			"template parse failed",
			"err", err,
		)
		http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, pageData)
	if err != nil {
		slog.Error(
			"template execute failed",
			"err", err,
		)
		http.Error(w, "uh-oh, something went wrong", http.StatusInternalServerError)
		return
	}
}

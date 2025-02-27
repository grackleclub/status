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

//go:embed static
var static embed.FS

var (
	queries     *db.Queries
	dbStr       = "file:status.db?cache=shared&mode=rw"
	portDefault = "8888"
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

	var targets = []string{
		"https://www.google.com",
		"https://api.grackle.club",
		// "https://fake.grackle.club",
	}

	go statusesForever(ctx, targets, 5*time.Second)

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
		"remote", r.RemoteAddr,
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

	type rslt struct {
		Time string
		Url  string
		Code int
		Rtt  int
	}
	var results []rslt
	for _, s := range statuses {
		results = append(results, rslt{
			Time: s.Ts.Format(time.RFC3339),
			Url:  s.Url,
			Code: int(s.StatusCode),
			Rtt:  int(s.ResponseMs),
		})
	}

	tmpl, err := template.ParseFS(static, path.Join("static", "index.html"))
	if err != nil {
		slog.Error(
			"template parse failed",
			"err", err,
		)
		http.Error(w, "oops, something went wrong", http.StatusInternalServerError)
		return
	}
	for _, r := range results {
		slog.Debug("results", "r", r)
	}
	err = tmpl.Execute(w, results)
	if err != nil {
		slog.Error(
			"template execute failed",
			"err", err,
		)
		http.Error(w, "uh-oh, something went wrong", http.StatusInternalServerError)
		return
	}
}

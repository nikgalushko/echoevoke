package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nikgalushko/echoevoke/assets"
)

var args struct {
	init     bool
	port     int
	logLevel string
	dbFile   string
}

func init() {
	flag.BoolVar(&args.init, "init", false, "initialize SQL tables")
	flag.IntVar(&args.port, "port", 8080, "HTTP server port")
	flag.StringVar(&args.logLevel, "log-level", "info", "log level")
	flag.StringVar(&args.dbFile, "db-file", "echoevoke.db", "SQLite database file")

	flag.Usage = func() {
		fmt.Println("Usage: echoevoke [options]")
		flag.PrintDefaults()
	}

	flag.Parse()
}

func main() {
	if args.init {
		err := initSQL()
		if err != nil {
			_ = os.Remove(args.dbFile)
			fmt.Println("Failed to initialize SQL tables:", err)
			os.Exit(1)
		}
	} else {
		run()
	}
}

func initSQL() error {
	fmt.Println("Initializing SQL tables")

	db, err := sql.Open("sqlite3", args.dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	entries, err := assets.SQL.ReadDir("sql")
	if err != nil {
		return fmt.Errorf("failed to read sql directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			sqlBytes, err := assets.SQL.ReadFile(filepath.Join("sql", entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to read sql file: %w", err)
			}

			_, err = db.ExecContext(context.Background(), string(sqlBytes))
			if err != nil {
				return fmt.Errorf("failed to execute sql: %w", err)
			}
		}
	}

	return nil
}

func run() {
	fmt.Println("Running the server")
}

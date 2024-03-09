package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/nikgalushko/echoevoke/assets"
	"github.com/nikgalushko/echoevoke/internal/storage"
	"github.com/nikgalushko/echoevoke/internal/storage/disk"
)

var args struct {
	port     int
	logLevel string
	dbFile   string
}

func init() {
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
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initDB(db *sql.DB) error {
	fmt.Println("Initializing SQL tables")

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

func run() error {
	db, err := sql.Open("sqlite3", args.dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	err = initDB(db)
	if err != nil {
		_ = os.Remove(args.dbFile)
		return fmt.Errorf("failed to initialize SQL tables: %w", err)
	}

	fmt.Println("Running the server")

	s := NewServer(disk.NewChannelRegistry(db))

	err = http.ListenAndServe(fmt.Sprintf(":%d", args.port), s)
	if err != nil {
		return fmt.Errorf("failed to start the server: %w", err)
	}

	return nil
}

type Server struct {
	registry storage.ChannelsRegistry
	mux      *chi.Mux
}

func NewServer(registry storage.ChannelsRegistry) *Server {
	s := &Server{
		registry: registry,
		mux:      chi.NewRouter(),
	}

	s.routes()

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.Use(middleware.RequestID)
	s.mux.Use(middleware.Logger)
	s.mux.Use(middleware.Recoverer)

	s.mux.Route("/channel", func(r chi.Router) {
		r.Post("/register", s.handleChannelRegistration())
	})

	static, err := fs.Sub(assets.HTML, "html")
	if err != nil {
		log.Fatal("failed to read html directory:", err)
	}

	s.mux.Get("/", http.FileServer(http.FS(static)).ServeHTTP)
}

func (s *Server) handleChannelRegistration() http.HandlerFunc {
	type request struct {
		ChannelID string `json:"channel_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		err = s.registry.RegisterChannel(r.Context(), req.ChannelID)
		if err != nil {
			log.Println("[ERROR] handle channel registration:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

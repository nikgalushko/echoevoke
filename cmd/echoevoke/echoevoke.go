package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron/v3"

	"github.com/nikgalushko/echoevoke/assets"
	"github.com/nikgalushko/echoevoke/internal/scrapper"
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

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
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
	startAt := time.Now()
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
	posts := disk.NewPostsStorage(db)
	images := disk.NewImagesStorage(db)

	scrp := scrapper.New(posts, scrapper.NewImageDownloader(images))

	c := cron.New(cron.WithSeconds())
	c.AddFunc("0 * * * *", func() {
		dir := filepath.Join("./", strconv.FormatInt(time.Now().Unix(), 10))

		channels, err := s.registry.AllChannels(context.Background())
		if err != nil {
			slog.Error("failed to get all channels to scan", slog.Any("err", err))
			return
		}

		for _, ch := range channels {
			rootDir := filepath.Join(dir, ch)
			err := os.MkdirAll(rootDir, 0755)
			if err != nil {
				slog.Error(" failed to create the channel directory", slog.Any("err", err))
				continue
			}

			posts, err := posts.GetPosts(context.Background(), ch, startAt, time.Now())
			if err != nil {
				slog.Error("failed to get posts from", slog.String("value", ch), slog.Any("err", err))
				continue
			}

			for _, p := range posts {
				err = os.WriteFile(filepath.Join(rootDir, fmt.Sprintf("%d.md", p.ID)), []byte(p.Message), 0644)
				if err != nil {
					slog.Error(" failed to write the post file from", slog.String("value", ch), slog.Any("err", err))
				}
			}
		}
	})

	c.AddFunc("*/10 * * * *", func() {
		channels, err := s.registry.AllChannels(context.Background())
		if err != nil {
			slog.Error("failed to get all channels", slog.Any("err", err))
			return
		}

		for _, ch := range channels {
			err = scrp.Scrape(context.TODO(), ch)
			if err != nil {
				slog.Error("failed to scrape", slog.String("channel", ch), slog.Any("err", err))
			}
		}
	})
	c.Start()

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
		slog.Error("failed to read html directory", slog.Any("err", err))
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
			slog.Error("handle channel registration", slog.String("value", req.ChannelID), slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

package server

import (
	"befw/internal/logger"
	"befw/internal/postgres"
	"befw/internal/storage"

	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HTTPServer struct {
	Server    *http.Server
	chiRouter chi.Router
	Storage   storage.Storage
	KeySign   []byte
}

func NewServer(ctx context.Context, cfg *Config) *HTTPServer {
	httpserver := HTTPServer{
		chiRouter: chi.NewRouter(),
		KeySign:   []byte(cfg.KeySign),
	}

	if cfg.DatabaseDSN != "" {
		var db *pgxpool.Pool

		poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
		if err != nil {
			logger.Fatal("unable to parse database dsn", err)
		}

		db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			logger.Fatal("unable to create connection pool", err)
		}

		//httpserver.Storage = storage1.NewPostgresDB(db)
		httpserver.Storage = postgres.NewPostgresDB(db)
		err = httpserver.Storage.DB().InitDB()
		if err != nil {
			logger.Fatal("", err)
		}
	}

	return &httpserver
}

func (s *HTTPServer) ListenAndServe(addr string) {

	// middleware gzip response
	s.chiRouter.Use(middleware.Compress(5, "text/html", "application/json"))

	// middleware unzip request
	s.chiRouter.Use(unzipBodyHandler)

	s.chiRouter.Get("/", s.defaultHandler)
	s.chiRouter.Post("/", s.defaultHandler)
	s.chiRouter.Post("/api/v1/rule", s.getRule)
	s.chiRouter.Post("/api/v1/rule/add", s.addRule)
	s.chiRouter.Post("/api/v1/rule/delete", s.deleteRule)
	s.chiRouter.Post("/api/v1/rule/stat", s.updateRuleStat)

	s.Server = &http.Server{
		Addr:    addr,
		Handler: s.chiRouter,
	}

	if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("", err)
	}
}

func (s *HTTPServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed", err)
	}

	if err := s.Storage.DB().Close(); err != nil {
		logger.Fatal("Server storage close is failed", err)
	}
}

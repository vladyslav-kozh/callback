package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evt/callback/config"
	"github.com/evt/callback/internal/handlers/callbackhandler"
	"github.com/evt/callback/internal/pg"
	"github.com/evt/callback/internal/repositories/objectrepo"
	"github.com/evt/callback/internal/services/objectservice"
	"github.com/evt/callback/internal/services/testerservice"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultHTTPAddr = ":8080"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	defaultCtx := context.Background()
	cfg := config.Get()

	pgDB, err := pg.Dial()
	if err != nil {
		return fmt.Errorf("Failed to connect to PostgreSQL: %w", err)
	}

	if pgDB != nil {
		log.Println("Running PostgreSQL migrations")
		if err := runPgMigrations(cfg.PgMigrationsPath, cfg.PgURL); err != nil {
			return fmt.Errorf("Failed to run PostgreSQL migrations: %w", err)
		}
	}

	objectRepo := objectrepo.New(pgDB)

	go func() {
		if err := objectRepo.CleanExpiredObjects(defaultCtx); err != nil {
			log.Fatalf("Error cleaning expired objects: %v", err)
		}
	}()

	objectService := objectservice.New(objectRepo)
	testerService := testerservice.New(time.Second * 60)
	callbackHandler := callbackhandler.New(objectService, testerService)

	http.HandleFunc("/callback", callbackHandler.HandleCallbackRequest)

	httpAddr := cfg.HTTPAddr
	if httpAddr == "" {
		httpAddr = defaultHTTPAddr
	}

	log.Printf("Running HTTP server on %s\n", httpAddr)

	go func() {
		if err := http.ListenAndServe(httpAddr, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down gracefully")

	return nil
}

func runPgMigrations(migrationsPath, pgURL string) error {
	if migrationsPath == "" {
		return nil
	}

	if pgURL == "" {
		return errors.New("No PostgreSQL URL provided")
	}

	m, err := migrate.New(migrationsPath, pgURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

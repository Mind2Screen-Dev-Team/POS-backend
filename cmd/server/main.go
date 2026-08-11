package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/config"
	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/handler"
	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/repository"
)

func main() {
	logger := log.New(os.Stderr, "[pos-backend] ", log.LstdFlags|log.Lmsgprefix)

	// Load configuration from environment.
	cfg := config.Load()
	logger.Printf("starting in %s mode", cfg.AppEnv)

	// Connect to PostgreSQL.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := repository.New(ctx, cfg.DSN(), cfg.DBMaxConns)
	if err != nil {
		logger.Printf("WARNING: failed to connect to database: %v — starting without DB", err)
		db = nil
	} else {
		defer db.Close()
		logger.Println("database connected")
	}

	// Build router and server.
	router := handler.NewRouter(db)
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	errCh := make(chan error, 1)
	go func() {
		logger.Printf("listening on %s", cfg.Addr())
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Printf("received signal %v — shutting down", sig)
	case err := <-errCh:
		logger.Fatalf("server error: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("shutdown error: %v", err)
	}
	logger.Println("server stopped")
}

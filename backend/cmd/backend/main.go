package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rudraksh121a/BookStore/internal/config"
	"github.com/Rudraksh121a/BookStore/internal/http/handlers/books"
)

func main() {
	// fmt.Println("Hello, Backend!")
	//load config

	cfg := config.MustLoad()

	//db setup

	//setup router
	router := http.NewServeMux()

	router.HandleFunc("GET /api/health", books.New())
	router.HandleFunc("POST /api/users/register", books.Register(cfg))
	router.HandleFunc("POST /api/users/login", books.Login(cfg))

	//start server

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}
	slog.Info("server started", slog.String("Address", cfg.Addr))
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()
	<-done

	slog.Info("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server", slog.String("error", err.Error()))
	}
	slog.Info("Server Shutdown successfully")
}

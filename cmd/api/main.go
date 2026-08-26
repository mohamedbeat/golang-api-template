package main

import (
	"context"
	"fmt"
	"golang-api-template/internal/auth"
	"golang-api-template/internal/config"
	"golang-api-template/internal/database"
	"golang-api-template/internal/handler"
	"golang-api-template/internal/users"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoadConfig()
	err := database.Connect(cfg.Db.Host, cfg.Db.Port, cfg.Db.User, cfg.Db.Password, cfg.Db.DbName)
	if err != nil {
		panic(err)
	}

	db := database.GetDBConnection()

	//////////////////////////////
	// Initializing repos
	//////////////////////////////
	userRepo := users.NewUserRepository(db)
	sessionRepo := auth.NewSessionRepo(db)

	//////////////////////////////
	// Initializing services
	//////////////////////////////
	userService := users.NewUserService(userRepo)
	authService := auth.NewAuthService(cfg, sessionRepo, userRepo)

	//////////////////////////////
	// Initializing handlers
	//////////////////////////////
	userHandler := users.NewUserHandler(userService)
	authHandler := auth.NewAuthHandler(authService, userService)

	r := handler.NewRouter(userHandler, authHandler, authService)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("starting the server on port %v \n", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("could not listen on %d: %v\n", cfg.App.Port, err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("shutting down server...")
	db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	fmt.Println("server exited properly")
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BilyHakim/go-walet/config"
	"github.com/BilyHakim/go-walet/routes"
	"github.com/BilyHakim/go-walet/worker"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("warning: .env file not found, using env variables")
	}

	db := config.InitDB()
	rmq := config.InitRabbitMQ()

	transferWorker := worker.NewTransferWorker(db, rmq)
	transferWorker.Start()

	r := gin.Default()
	routes.SetupRoutes(r, db, rmq)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server starting on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	transferWorker.Stop()

	rmq.Close()

	log.Println("Server exited properly")
}

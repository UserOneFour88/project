package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pipelineapp/internal/auth"
	"pipelineapp/internal/config"
	"pipelineapp/internal/httpapi"
	"pipelineapp/internal/store"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Close()

	repo := store.NewRepo(db)
	jwtSvc := auth.NewService(cfg.JWTIssuer, cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.AccessTTL, cfg.RefreshTTL)
	api := httpapi.New(repo, jwtSvc)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("rest api listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
	log.Println("shutdown complete")
}


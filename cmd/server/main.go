package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/api"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/auth"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/config"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/store"
	syncpkg "github.com/XimilalaXiang/ReferenceAnswerRSS/internal/sync"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/xinzhi"
)

//go:embed all:dist
var webFS embed.FS

func main() {
	cfg := config.Load()

	if cfg.XinzhiToken == "" {
		log.Fatal("XINZHI_TOKEN environment variable is required")
	}

	st, err := store.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer st.Close()

	if err := st.EnsureAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	jwtAuth := auth.New(cfg.JWTSecret)
	xinzhiClient := xinzhi.NewClient(cfg.XinzhiAPIBase, cfg.XinzhiToken)
	syncService := syncpkg.New(xinzhiClient, st, cfg.AuthorID, cfg.SyncInterval)

	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Port)
	if v := os.Getenv("BASE_URL"); v != "" {
		baseURL = v
	}

	mux := http.NewServeMux()

	handler := api.NewHandler(st, jwtAuth, syncService, baseURL)
	handler.RegisterRoutes(mux)

	distFS, err := fs.Sub(webFS, "dist")
	if err != nil {
		log.Fatalf("Failed to get embedded FS: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" || path == "/index.html" {
			r.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r)
			return
		}

		f, err := distFS.Open(path[1:])
		if err != nil {
			r.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})

	syncService.Start()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: corsMiddleware(mux),
	}

	go func() {
		log.Printf("Server starting on :%d", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	syncService.Stop()
	server.Close()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

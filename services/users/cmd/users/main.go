package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"refurbished-marketplace/services/users/internal/database"
	"refurbished-marketplace/services/users/internal/handlers"
	"refurbished-marketplace/services/users/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	queries := database.New(db)
	jwtSecret := os.Getenv("JWT_SECRET")
	cfg := service.DefaultConfig(jwtSecret)
	if err := service.ValidateConfig(cfg); err != nil {
		log.Fatalf("auth config: %v", err)
	}

	svc := service.New(queries, cfg)
	h := handlers.New(svc)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("starting users service on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/cart/internal/migration"
	"github.com/redis/go-redis/v9"
)

func main() {
	sourceAddr := flag.String("source", "", "source Redis address")
	destinationAddr := flag.String("destination", "", "destination Redis address")
	flag.Parse()
	if *sourceAddr == "" || *destinationAddr == "" {
		log.Fatal("source and destination addresses are required; stop cart writes before copying")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	source := redis.NewClient(&redis.Options{Addr: *sourceAddr})
	defer func() { _ = source.Close() }()
	destination := redis.NewClient(&redis.Options{Addr: *destinationAddr})
	defer func() { _ = destination.Close() }()
	copied, err := migration.CopyCarts(ctx, source, destination)
	if err != nil {
		log.Fatalf("copied %d carts before failure: %v", copied, err)
	}
	log.Printf("copied %d carts with their remaining TTLs", copied)
}

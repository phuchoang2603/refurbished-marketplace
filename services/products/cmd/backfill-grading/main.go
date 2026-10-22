package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/catalog"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
)

func main() {
	apply := flag.Bool("apply", false, "write missing grading fields (default: dry run)")
	flag.Parse()
	cfg := service.LoadConfig()
	if cfg.MongoURI == "" {
		fmt.Fprintln(os.Stderr, "Mongo connection is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	store, err := catalog.Open(ctx, cfg.MongoURI)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open catalog failed")
		os.Exit(1)
	}
	defer store.Close(context.Background())
	result, err := store.BackfillGrading(ctx, *apply)
	if err != nil {
		fmt.Fprintln(os.Stderr, "backfill failed")
		os.Exit(1)
	}
	fmt.Printf("apply=%t category=%d condition=%d\n", *apply, result.Category, result.Condition)
}

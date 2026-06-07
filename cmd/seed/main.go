package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	outPath := flag.String("out", "docs/SEED_DATA.xlsx", "Excel output path")
	skipClean := flag.Bool("skip-clean", false, "Do not remove existing demo data first")
	flag.Parse()

	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required (set in .env or environment)")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if !*skipClean {
		log.Println("Cleaning previous demo data (@demo.go-connect.local)...")
		if err := cleanupDemoData(ctx, db); err != nil {
			log.Fatalf("cleanup: %v", err)
		}
	}

	catalog := &Catalog{}
	s := &seeder{db: db, catalog: catalog, now: time.Now().UTC()}

	log.Println("Seeding demo data...")
	if err := s.run(ctx); err != nil {
		log.Fatalf("seed: %v", err)
	}

	if err := os.MkdirAll("docs", 0o755); err != nil {
		log.Fatalf("mkdir docs: %v", err)
	}
	if err := exportExcel(catalog, *outPath); err != nil {
		log.Fatalf("excel: %v", err)
	}

	fmt.Println()
	fmt.Println("Demo seed complete.")
	fmt.Printf("  Excel reference: %s\n", *outPath)
	fmt.Printf("  Password (all demo users): %s\n", demoPassword)
	fmt.Println()
	fmt.Println("  Admin:    admin@demo.go-connect.local     → http://localhost:3001/login")
	fmt.Println("  Customer: alice@demo.go-connect.local    → http://localhost:3002/login")
	fmt.Println("  Employee: karim@demo.go-connect.local    → http://localhost:3003/login")
}

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"blog.simoni.dev/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

var (
	//go:embed db/migrations
	migrationsFs embed.FS
)

func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("running with %d CPUs\n", nuCPU)
}

func main() {
	godotenv.Load()

	ConfigRuntime()

	tz := os.Getenv("APP_TZ")
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			log.Fatal("invalid APP_TZ: ", err)
		}
		time.Local = loc
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DSN"))
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)
	goose.SetBaseFS(migrationsFs)
	goose.SetDialect("postgres")
	if err := goose.Up(sqlDB, "db/migrations"); err != nil {
		log.Fatal("migration failed: ", err)
	}

	engine, err := server.NewServer(pool)
	if err != nil {
		log.Fatal("failed to create server: ", err)
	}

	if err = engine.Run(":8080"); err != nil {
		log.Fatal("failed to run server: ", err)
	}
}

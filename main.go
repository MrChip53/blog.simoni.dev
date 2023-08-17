package main

import (
	"blog.simoni.dev/models"
	"blog.simoni.dev/server"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"runtime"
)

func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("running with %d CPUs\n", nuCPU)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ConfigRuntime()

	db, err := gorm.Open(
		mysql.Open(os.Getenv("DSN")),
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}

	err = db.AutoMigrate(&models.BlogPost{}, &models.Tag{})
	if err != nil {
		log.Fatal("failed to migrate db", err)
	}

	engine, err := server.NewServer(db)
	if err != nil {
		log.Fatal("failed to create server", err)
	}

	if err = engine.Run(":8080"); err != nil {
		log.Fatal("failed to run server", err)
	}
}

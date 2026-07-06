package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"blog.simoni.dev/models"
	"blog.simoni.dev/server"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	db, err := gorm.Open(
		postgres.Open(os.Getenv("DSN")),
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})

	if err != nil {
		log.Fatal("failed to open db connection ", err)
	}

	err = db.AutoMigrate(&models.BlogPost{}, &models.Tag{}, &models.User{}, &models.Comment{})
	if err != nil {
		log.Fatal("failed to migrate db: ", err)
	}

	engine, err := server.NewServer(db)
	if err != nil {
		log.Fatal("failed to create server: ", err)
	}

	if err = engine.Run(":8080"); err != nil {
		log.Fatal("failed to run server: ", err)
	}
}

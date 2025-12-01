package main

import (
	"fmt"
	"log"
	"runtime"

	"blog.simoni.dev/models"
	"blog.simoni.dev/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
		sqlite.Open("blog.db"),
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

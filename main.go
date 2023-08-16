package main

import (
	"blog.simoni.dev/models"
	"blog.simoni.dev/server"
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"html/template"
	"log"
	"os"
	"path"
	"runtime"
	"time"
)

func getSlug(post models.BlogPost) string {
	return fmt.Sprintf("/post/%02d/%02d/%d/%s", post.CreatedAt.Month(), post.CreatedAt.Day(), post.CreatedAt.Year(), post.Slug)
}

func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}

	return s
}

func formatAsDateTime(t time.Time) string {
	year, month, day := t.Date()
	dateString := fmt.Sprintf("%d/%02d/%02d", year, month, day)

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		fmt.Println("Error loading time zone:", err)
		return dateString
	}

	chicagoTime := t.In(loc)

	timeFormat := "01/02/2006 3:04 PM"

	timeString := chicagoTime.Format(timeFormat)

	return timeString
}

func renderer(templatePath string) multitemplate.Renderer {
	funcMap := template.FuncMap{
		"formatAsDateTime": formatAsDateTime,
		"getSlug":          getSlug,
		"truncateString":   truncateString,
	}

	basePath := path.Join(templatePath, "base.html")

	r := multitemplate.NewRenderer()
	r.AddFromFilesFuncs("index", funcMap, basePath, path.Join(templatePath, "index.html"))
	r.AddFromFilesFuncs("post", funcMap, basePath, path.Join(templatePath, "post.html"))

	// Error pages
	r.AddFromFilesFuncs("notFound", funcMap, basePath, path.Join(templatePath, "errors/404.html"))

	return r
}

func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("Running with %d CPUs\n", nuCPU)
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
		log.Fatal("Failed to migrate db", err)
	}

	controller := server.NewController(db)

	//db.Create(&BlogPost{
	//	Title:   "Under Development",
	//	Author:  "mrchip53",
	//	Slug:    "under-development",
	//	Content: "This blog is currently under development. I am making it using Golang and Gin. Tailwind CSS is being used for CSS and GORM for the ORM. If I want to include rich user interactions I will probably use HTMX.",
	//	Tags: []Tag{
	//		{
	//			Name: "programming",
	//		},
	//		{
	//			Name: "golang",
	//		},
	//	},
	//})

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = log.Writer()

	router := gin.Default()

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal("failed to SetTrustedProxies:", err)
	}

	router.Static("/css", "css")
	router.HTMLRender = renderer("./templates")

	router.NoRoute(controller.HandleNotFound)

	router.GET("/", controller.HandleIndex)
	router.GET("/post/:month/:day/:year/:slug", controller.HandlePost)
	router.GET("/tag/:tag", controller.HandleTag)
	router.GET("/hp", controller.HandleHealth)

	if err = router.Run(":8080"); err != nil {
		log.Fatal("failed to run router:", err)
	}
}

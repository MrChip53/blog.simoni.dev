package main

import (
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
)

var db *gorm.DB

type BlogPost struct {
	gorm.Model
	Title   string
	Author  string
	Slug    string `gorm:"type:varchar(100);unique_index"`
	Content string `gorm:"type:text"`
	Tags    []Tag  `gorm:"many2many:blog_post_tags;"`
}

type Tag struct {
	gorm.Model
	Name string
}

func getSlug(post BlogPost) string {
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

	return r
}

func Index(ctx *gin.Context) {
	var posts []BlogPost
	if err := db.Preload("Tags").Order("created_at DESC").Limit(10).Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.HTML(200, "index", gin.H{
		"posts":   posts,
		"noPosts": len(posts) == 0,
	})
}

func HandlePost(ctx *gin.Context) {
	month := ctx.Param("month")
	day := ctx.Param("day")
	year := ctx.Param("year")
	slug := ctx.Param("slug")

	var post BlogPost
	if err := db.Preload("Tags").Where("day(created_at) = ? AND month(created_at) = ? AND year(created_at) = ? AND slug = ?", day, month, year, slug).First(&post).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.HTML(200, "post", gin.H{
		"post": post,
	})
}

func HandleTag(ctx *gin.Context) {
	tag := ctx.Param("tag")

	var posts []BlogPost
	if err := db.Preload("Tags", func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ?", tag)
	}).Order("created_at DESC").Limit(10).Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	for i, post := range posts {
		var tags []Tag
		if err := db.Model(&post).Association("Tags").Find(&tags); err != nil {
			log.Println("Index failed to get tags:", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		posts[i].Tags = tags
	}

	ctx.HTML(200, "index", gin.H{
		"posts":   posts,
		"noPosts": len(posts) == 0,
	})
}

func Health(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
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

	db, err = gorm.Open(
		mysql.Open(os.Getenv("DSN")),
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}

	err = db.AutoMigrate(&BlogPost{}, &Tag{})
	if err != nil {
		log.Fatal("Failed to migrate db", err)
	}

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

	router.GET("/", Index)
	router.GET("/post/:month/:day/:year/:slug", HandlePost)
	router.GET("/tag/:tag", HandleTag)
	router.GET("/hp", Health)

	if err = router.Run(":8080"); err != nil {
		log.Fatal("failed to run router:", err)
	}

	//fs := http.FileServer(http.Dir("./public"))
	//http.Handle("/", fs)
	//
	//log.Print("Listening on :3000...")
	//err = http.ListenAndServe(":3000", nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
}

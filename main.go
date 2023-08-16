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

	// Error pages
	r.AddFromFilesFuncs("notFound", funcMap, basePath, path.Join(templatePath, "errors/404.html"))

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
		"title":   "mrchip53's blog",
		"posts":   posts,
		"noPosts": len(posts) == 0,
	})
}

func Handle404(ctx *gin.Context) {
	ctx.HTML(404, "notFound", gin.H{
		"title": "Content Not Found",
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
		ctx.HTML(404, "notFound", gin.H{
			"title": "Content Not Found",
		})
		return
	}

	ctx.HTML(200, "post", gin.H{
		"title": post.Title,
		"post":  post,
	})
}

func HandleTag(ctx *gin.Context) {
	tag := ctx.Param("tag")

	var postIds []uint
	// Raw sql query to pull post ids
	if err := db.Raw("SELECT blog_post_id FROM blog_post_tags WHERE tag_id = (SELECT id FROM tags WHERE name = ?)", tag).Scan(&postIds).Error; err != nil {
		log.Println("Tag failed to get post ids:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var posts []BlogPost
	if err := db.Preload("Tags").Where("id IN ?", postIds).Find(&posts).Error; err != nil {
		log.Println("Tag failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	//for i, post := range posts {
	//	var tags []Tag
	//	if err := db.Model(&post).Association("Tags").Find(&tags); err != nil {
	//		log.Println("Tag failed to get tags:", err)
	//		ctx.AbortWithStatus(http.StatusInternalServerError)
	//		return
	//	}
	//	posts[i].Tags = tags
	//}

	ctx.HTML(200, "index", gin.H{
		"title":   fmt.Sprintf("Posts tagged with %s", tag),
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

	router.NoRoute()

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

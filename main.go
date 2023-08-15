package main

import (
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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
	Content []Section `gorm:"foreignKey:BlogPostID"`
	Tags    []Tag     `gorm:"many2many:blog_post_tags;"`
}

type Section struct {
	gorm.Model
	Header     string
	Content    string
	BlogPostID uint
}

type Tag struct {
	gorm.Model
	Name string
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d/%02d/%02d", year, month, day)
}

func formatAsRfc3339String(t time.Time) string {
	return t.Format(time.RFC3339)
}

func renderer(templatePath string) multitemplate.Renderer {
	funcMap := template.FuncMap{
		"formatAsDate":          formatAsDate,
		"formatAsRfc3339String": formatAsRfc3339String,
	}

	basePath := path.Join(templatePath, "base.html")

	r := multitemplate.NewRenderer()
	r.AddFromFilesFuncs("index", funcMap, basePath, path.Join(templatePath, "index.html"))

	return r
}

func Index(ctx *gin.Context) {
	var posts []BlogPost
	if err := db.Preload("Content").Preload("Tags").Limit(10).Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.HTML(200, "index", gin.H{
		"posts":   posts,
		"noPosts": len(posts) == 0,
	})
}

func ConfigRuntime() {
	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	fmt.Printf("Running with %d CPUs\n", nuCPU)
}

func main() {
	ConfigRuntime()

	var err error
	db, err = gorm.Open(
		mysql.Open(os.Getenv("DSN")),
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}

	err = db.AutoMigrate(&BlogPost{}, &Section{}, &Tag{})
	if err != nil {
		log.Fatal("Failed to migrate db", err)
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = log.Writer()

	router := gin.Default()

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal("failed to SetTrustedProxies:", err)
	}

	router.Static("/css", "css")
	router.HTMLRender = renderer("./templates")

	router.GET("/", Index)

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

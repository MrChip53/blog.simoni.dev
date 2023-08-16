package server

import (
	"blog.simoni.dev/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type Controller struct {
	Db *gorm.DB
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{Db: db}
}

func (c *Controller) HandleIndex(ctx *gin.Context) {
	var posts []models.BlogPost
	if err := c.Db.Preload("Tags").Order("created_at DESC").Limit(10).Find(&posts).Error; err != nil {
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

func (c *Controller) HandlePost(ctx *gin.Context) {
	month := ctx.Param("month")
	day := ctx.Param("day")
	year := ctx.Param("year")
	slug := ctx.Param("slug")

	var post models.BlogPost
	if err := c.Db.Preload("Tags").Where("day(created_at) = ? AND month(created_at) = ? AND year(created_at) = ? AND slug = ?", day, month, year, slug).First(&post).Error; err != nil {
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

func (c *Controller) HandleTag(ctx *gin.Context) {
	tag := ctx.Param("tag")

	var postIds []uint
	if err := c.Db.Raw("SELECT blog_post_id FROM blog_post_tags WHERE tag_id = (SELECT id FROM tags WHERE name = ?)", tag).Scan(&postIds).Error; err != nil {
		log.Println("Tag failed to get post ids:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var posts []models.BlogPost
	if err := c.Db.Preload("Tags").Where("id IN ?", postIds).Order("created_at DESC").Find(&posts).Error; err != nil {
		log.Println("Tag failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.HTML(200, "index", gin.H{
		"title":   fmt.Sprintf("Posts tagged with %s", tag),
		"posts":   posts,
		"noPosts": len(posts) == 0,
	})
}

func (c *Controller) HandleNotFound(ctx *gin.Context) {
	ctx.HTML(404, "notFound", gin.H{
		"title": "Content Not Found",
	})
}

func (c *Controller) HandleInternalServerError(ctx *gin.Context) {
	ctx.HTML(500, "notFound", gin.H{
		"title": "Internal Server Error",
	})
}

func (c *Controller) HandleHealth(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "ok",
	})
}

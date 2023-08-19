package server

import (
	"blog.simoni.dev/models"
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
	"path"
)

type Router struct {
	Db *gorm.DB
}

func createRenderer(templatePath string) multitemplate.Renderer {
	funcMap := template.FuncMap{
		"formatAsDateTime": formatAsDateTime,
		"getSlug":          getSlug,
		"truncateString":   truncateString,
	}

	basePath := path.Join(templatePath, "base.html")

	r := multitemplate.NewRenderer()

	// Regular pages
	r.AddFromFilesFuncs("index", funcMap, basePath, path.Join(templatePath, "index.html"))
	r.AddFromFilesFuncs("post", funcMap, basePath, path.Join(templatePath, "post.html"))

	// Admin pages
	r.AddFromFilesFuncs("adminLogin", funcMap, basePath, path.Join(templatePath, "admin/login.html"))
	r.AddFromFilesFuncs("adminDashboard", funcMap, basePath, path.Join(templatePath, "admin/dashboard.html"))

	// Error pages
	r.AddFromFilesFuncs("notFound", funcMap, basePath, path.Join(templatePath, "errors/404.html"))

	return r
}

func NewRouter(db *gorm.DB) *Router {
	return &Router{Db: db}
}

func (r *Router) HandleIndex(ctx *gin.Context) {
	var posts []models.BlogPost
	if err := r.Db.Preload("Tags").Order("created_at DESC").Limit(10).Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.HTML(200, "index", addHXRequest(ctx, gin.H{
		"title":   "mrchip53's blog",
		"posts":   posts,
		"noPosts": len(posts) == 0,
	}))
}

func (r *Router) HandlePost(ctx *gin.Context) {
	month := ctx.Param("month")
	day := ctx.Param("day")
	year := ctx.Param("year")
	slug := ctx.Param("slug")

	var post models.BlogPost
	if err := r.Db.Preload("Tags").Where("day(created_at) = ? AND month(created_at) = ? AND year(created_at) = ? AND slug = ?", day, month, year, slug).First(&post).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.HTML(404, "notFound", addHXRequest(ctx, gin.H{
			"title": "Content Not Found",
		}))
		return
	}

	ctx.HTML(200, "post", addHXRequest(ctx, gin.H{
		"title": post.Title,
		"post":  post,
	}))
}

func (r *Router) HandleTag(ctx *gin.Context) {
	tag := ctx.Param("tag")

	var postIds []uint
	if err := r.Db.Raw(
		"SELECT blog_post_id FROM blog_post_tags WHERE tag_id = (SELECT id FROM tags WHERE name = ?)",
		tag,
	).Scan(&postIds).Error; err != nil {
		log.Println("Tag failed to get post ids:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var posts []models.BlogPost
	if err := r.Db.Preload("Tags").Where("id IN ?", postIds).Order("created_at DESC").Find(&posts).Error; err != nil {
		log.Println("Tag failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.HTML(200, "index", addHXRequest(ctx, gin.H{
		"title":   fmt.Sprintf("Posts tagged with %s", tag),
		"posts":   posts,
		"noPosts": len(posts) == 0,
	}))
}

func (r *Router) HandleNotFound(ctx *gin.Context) {
	ctx.HTML(404, "notFound", addHXRequest(ctx, gin.H{
		"title": "Content Not Found",
	}))
}

func (r *Router) HandleInternalServerError(ctx *gin.Context) {
	ctx.HTML(500, "notFound", addHXRequest(ctx, gin.H{
		"title": "Internal Server Error",
	}))
}

func (r *Router) HandleHealth(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "ok",
	})
}

func (r *Router) HandleAdminLoginRequest(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	var user models.User
	if err := r.Db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Println("AdminLogin failed to get user:", err)
		ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
			"title": "Admin Login",
			"path":  ctx.Request.URL.Path,
			"error": "Invalid username or password",
		}))
		return
	}

	if match, err := user.VerifyPassword(password); err != nil {
		log.Println("AdminLogin failed to verify password:", err)
		ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
			"title": "Admin Login",
			"path":  ctx.Request.URL.Path,
			"error": "Invalid username or password",
		}))
		return
	} else if !match {
		ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
			"title": "Admin Login",
			"path":  ctx.Request.URL.Path,
			"error": "Invalid username or password",
		}))
		return
	}

	ctx.SetCookie("token", "", 60, "/", "blog.simoni.dev", true, true)
	ctx.SetCookie("refreshToken", "", 60*60*3, "/", "blog.simoni.dev", true, true)
	ctx.Redirect(302, adminRoute)
}

func (r *Router) HandleAdminDashboard(ctx *gin.Context) {
	ctx.HTML(200, "adminDashboard", addHXRequest(ctx, gin.H{
		"title": "Admin Dashboard",
	}))
}

func HandleAdminLogin(ctx *gin.Context) {
	ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
		"title": "Admin Login",
		"path":  ctx.Request.URL.Path,
	}))
}

func addHXRequest(ctx *gin.Context, h gin.H) gin.H {
	hxRequest, exists := ctx.Get("isHXRequest")
	h["isHXRequest"] = exists && hxRequest.(bool)
	return h
}

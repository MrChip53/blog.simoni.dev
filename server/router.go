package server

import (
	"blog.simoni.dev/auth"
	"blog.simoni.dev/models"
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	// Components
	r.AddFromFilesFuncs("toast", funcMap, path.Join(templatePath, "components/toast.html"))

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
	redirectPath := ctx.PostForm("redirect")

	var user models.User
	if err := r.Db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Println("AdminLogin failed to get user:", err)
		r.HandleError(ctx, "Invalid username or password", func(ctx *gin.Context) {
			ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
				"title": "Admin Login",
				"path":  ctx.Request.URL.Path,
				"error": "Invalid username or password",
			}))
		}, err)
		return
	}

	if match, err := user.VerifyPassword(password); err != nil {
		log.Println("AdminLogin failed to verify password:", err)
		r.HandleError(ctx, "Invalid username or password", func(ctx *gin.Context) {
			ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
				"title": "Admin Login",
				"path":  ctx.Request.URL.Path,
				"error": "Invalid username or password",
			}))
		}, err)
		return
	} else if !match {
		r.HandleError(ctx, "Invalid username or password", func(ctx *gin.Context) {
			ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
				"title": "Admin Login",
				"path":  ctx.Request.URL.Path,
				"error": "Invalid username or password",
			}))
		}, err)
		return
	}

	err := user.NewAuthTokens(ctx)
	if err != nil {
		log.Println("AdminLogin failed to generate tokens:", err)
		r.HandleError(ctx, "Invalid username or password", func(ctx *gin.Context) {
			ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
				"title": "Admin Login",
				"path":  ctx.Request.URL.Path,
				"error": "Something went wrong",
			}))
		}, err)
		return
	}

	ctx.Redirect(302, redirectPath)
}

func (r *Router) HandleAdminDashboard(ctx *gin.Context) {
	claims := ctx.MustGet("authToken").(*auth.JwtPayload)
	if claims == nil {
		ctx.Redirect(302, "/admin/login?redirect="+ctx.Request.URL.Path)
		ctx.Abort()
		return
	}

	ctx.HTML(200, "adminDashboard", addHXRequest(ctx, gin.H{
		"title":    "Admin Dashboard",
		"username": claims.Username,
	}))
}

func (r *Router) HandleAdminLogin(ctx *gin.Context) {
	redirect := ctx.Request.URL.Query().Get("redirect")

	if redirect == "" {
		redirect = adminRoute
	}

	ctx.HTML(200, "adminLogin", addHXRequest(ctx, gin.H{
		"title":    "Admin Login",
		"redirect": redirect,
	}))
}

func (r *Router) HandleError(ctx *gin.Context, message string, fn func(ctx *gin.Context), err error) {
	uuid := uuid.New().String()

	hxRequest, exists := ctx.Get("isHXRequest")
	if exists && hxRequest.(bool) {
		ctx.Header("HX-Retarget", "#toastContainer")
		ctx.Header("HX-Reswap", "beforeend")
		ctx.HTML(200, "toast", gin.H{
			"toastId": "toast-" + uuid,
			"toast":   message,
		})
		return
	}

	if fn != nil {
		fn(ctx)
	} else {
		ctx.HTML(500, "notFound", gin.H{
			"title": "Internal Server Error",
		})
	}
}

func addHXRequest(ctx *gin.Context, h gin.H) gin.H {
	hxRequest, exists := ctx.Get("isHXRequest")
	h["isHXRequest"] = exists && hxRequest.(bool)
	return h
}

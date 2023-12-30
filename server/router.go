package server

import (
	"blog.simoni.dev/auth"
	"blog.simoni.dev/models"
	"errors"
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
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
	r.AddFromFilesFuncs("comments", funcMap, basePath, path.Join(templatePath, "comments.html"))

	// Admin pages

	r.AddFromFilesFuncs("adminDashboard", funcMap, basePath, path.Join(templatePath, "admin/dashboard.html"))
	r.AddFromFilesFuncs("adminLogin", funcMap, basePath, path.Join(templatePath, "admin/login.html"))
	r.AddFromFilesFuncs("adminNewPost", funcMap, basePath, path.Join(templatePath, "admin/new-post.html"))
	r.AddFromFilesFuncs("postEdit", funcMap, basePath, path.Join(templatePath, "admin/edit.html"))

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

	ctx.Header("HX-Title", "mrchip53's blog")

	ctx.HTML(200, "index", addGenerics(ctx, gin.H{
		"title":   "mrchip53's blog",
		"posts":   posts,
		"noPosts": len(posts) == 0,
	}))
}

func (r *Router) HandleUser(ctx *gin.Context) {
	username := ctx.Param("username")
	var posts []models.BlogPost
	if err := r.Db.Preload("Tags").Where("author = ?", username).Order("created_at DESC").Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Header("HX-Title", username+"'s Page")

	ctx.HTML(200, "index", addGenerics(ctx, gin.H{
		"title":   username + "'s Page",
		"posts":   posts,
		"noPosts": len(posts) == 0,
	}))
}

func (r *Router) HandleComment(ctx *gin.Context) {
	if t, ok := ctx.Get("authToken"); !ok || t == nil {
		r.HandleError(ctx, "You must be logged in to comment", nil, nil)
		return
	}

	postId := ctx.Param("postId")
	author := ctx.PostForm("Username")
	comment := ctx.PostForm("comment")

	if len(comment) == 0 || len(author) == 0 {
		r.HandleError(ctx, "Author and comment cannot be empty", nil, nil)
		return
	}

	err := r.Db.Transaction(func(tx *gorm.DB) error {
		var blogPost models.BlogPost
		if err := tx.First(&blogPost, postId).Error; err != nil {
			return err
		}

		err := tx.Create(&models.Comment{
			BlogPostId: blogPost.ID,
			Author:     author,
			Comment:    comment,
		}).Error
		return err
	})
	if err != nil {
		r.HandleError(ctx, "Failed to create comment", nil, err)
		return
	}

	var comments []models.Comment
	r.Db.Where("blog_post_id = ?", postId).Order("created_at DESC").Find(&comments)

	ctx.Header("HX-Retarget", "#comments-"+postId)
	ctx.Header("HX-Reswap", "innerHTML")

	ctx.HTML(200, "comments", addGenerics(ctx, gin.H{
		"comments": comments,
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
		ctx.HTML(404, "notFound", addGenerics(ctx, gin.H{
			"title": "Content Not Found",
		}))
		return
	}

	var comments []models.Comment
	r.Db.Where("blog_post_id = ?", post.ID).Order("created_at DESC").Find(&comments)

	postHtml := parseMarkdown([]byte(post.Content))

	ctx.Header("HX-Title", post.Title)

	ctx.HTML(200, "post", addGenerics(ctx, gin.H{
		"title":       post.Title,
		"post":        post,
		"comments":    comments,
		"contentHtml": template.HTML(postHtml),
	}))
}

func (r *Router) HandlePostEdit(ctx *gin.Context) {
	postId := ctx.Param("postId")

	var post models.BlogPost
	if err := r.Db.Preload("Tags").Where("id = ?", postId).First(&post).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.HTML(404, "notFound", addGenerics(ctx, gin.H{
			"title": "Content Not Found",
		}))
		return
	}

	postHtml := parseMarkdown([]byte(post.Content))

	ctx.Header("HX-Title", "Editing "+post.Title)

	ctx.HTML(200, "postEdit", addGenerics(ctx, gin.H{
		"title":       post.Title,
		"post":        post,
		"contentHtml": template.HTML(postHtml),
	}))
}

func (r *Router) PostPostEdit(ctx *gin.Context) {
	postId := ctx.Param("postId")

	content := ctx.PostForm("content")

	if len(content) == 0 {
		r.HandleError(ctx, "Just delete the post instead.", nil, nil)
		return
	}

	err := r.Db.Model(&models.BlogPost{}).Where("id = ?", postId).Update("content", strings.TrimSpace(content)).Error
	if err != nil {
		r.HandleError(ctx, "Failed to update post.", nil, err)
		return
	}
	var post models.BlogPost
	if err = r.Db.Where("id = ?", postId).First(&post).Error; err != nil {
		r.HandleError(ctx, "Failed to load new content.", nil, err)
		return
	}
	location := fmt.Sprintf("{ \"path\": \"/post/%d/%d/%d/%s\", \"target\":\"#main-container\"}", post.CreatedAt.Month(), post.CreatedAt.Day(), post.CreatedAt.Year(), post.Slug)
	ctx.Header("HX-Location", location)
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

	ctx.Header("HX-Title", "Posts tagged with "+tag)

	ctx.HTML(200, "index", addGenerics(ctx, gin.H{
		"title":   fmt.Sprintf("Posts tagged with %s", tag),
		"posts":   posts,
		"noPosts": len(posts) == 0,
	}))
}

func (r *Router) HandleNotFound(ctx *gin.Context) {
	ctx.HTML(404, "notFound", addGenerics(ctx, gin.H{
		"title": "Content Not Found",
	}))
}

func (r *Router) HandleInternalServerError(ctx *gin.Context) {
	ctx.HTML(500, "notFound", addGenerics(ctx, gin.H{
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
			ctx.HTML(200, "adminLogin", addGenerics(ctx, gin.H{
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
			ctx.HTML(200, "adminLogin", addGenerics(ctx, gin.H{
				"title": "Admin Login",
				"path":  ctx.Request.URL.Path,
				"error": "Invalid username or password",
			}))
		}, err)
		return
	} else if !match {
		r.HandleError(ctx, "Invalid username or password", func(ctx *gin.Context) {
			ctx.HTML(200, "adminLogin", addGenerics(ctx, gin.H{
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
			ctx.HTML(200, "adminLogin", addGenerics(ctx, gin.H{
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

	ctx.HTML(200, "adminDashboard", addGenerics(ctx, gin.H{
		"title":    "Admin Dashboard",
		"username": claims.Username,
	}))
}

func (r *Router) HandleAdminGenerateMarkdown(ctx *gin.Context) {
	claims := ctx.MustGet("authToken").(*auth.JwtPayload)
	if claims == nil {
		ctx.Abort()
		return
	}
	md := ctx.PostForm("content")
	htmlBytes := parseMarkdown([]byte(strings.TrimSpace(md)))
	ctx.String(200, string(htmlBytes))
}

func (r *Router) HandleAdminLogin(ctx *gin.Context) {
	redirect := ctx.Request.URL.Query().Get("redirect")

	if redirect == "" {
		redirect = adminRoute
	}

	ctx.HTML(200, "adminLogin", addGenerics(ctx, gin.H{
		"title":    "Admin Login",
		"redirect": redirect,
	}))
}

func (r *Router) HandleAdminNewBlogPost(ctx *gin.Context) {
	ctx.HTML(200, "adminNewPost", addGenerics(ctx, gin.H{
		"title": "New Blog Post",
	}))
}

func (r *Router) HandleAdminNewBlogPostRequest(ctx *gin.Context) {
	title := ctx.PostForm("title")
	tags := strings.Split(ctx.PostForm("tags"), ",")
	content := ctx.PostForm("content")
	description := ctx.PostForm("description")
	slug := strings.ReplaceAll(title, " ", "-")

	jwt, exists := ctx.Get("authToken")
	if !exists {
		r.HandleError(ctx, "You are not logged in", nil, nil)
		return
	}
	author := jwt.(*auth.JwtPayload).Username

	tx := r.Db.Begin()
	newPost, err := models.NewBlogPost(tx, title, author, slug, strings.TrimSpace(content), description)
	if err != nil {
		tx.Rollback()
		r.HandleError(ctx, "Failed to create blog post", nil, err)
		return
	}

	for _, tag := range tags {
		tagModel, err := models.GetTag(tx, tag)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tagModel, err = models.NewTag(tx, tag)
				if err != nil {
					tx.Rollback()
					r.HandleError(ctx, "Failed to create blog post", nil, err)
					return
				}
			} else {
				tx.Rollback()
				r.HandleError(ctx, "Failed to create blog post", nil, err)
				return
			}
		}
		// TODO why is this needed?
		time.Sleep(10 * time.Millisecond)
		newPost.AddTag(tagModel)
	}

	err = newPost.UpdateTags(tx)
	if err != nil {
		tx.Rollback()
		r.HandleError(ctx, "Failed to create blog post", nil, err)
		return
	}

	if err = tx.Commit().Error; err != nil {
		r.HandleError(ctx, "Failed to create blog post", nil, err)
		return
	}

	ctx.Redirect(302, adminRoute)
}

func (r *Router) HandleAdminPostsDelete(ctx *gin.Context) {
	err := r.Db.Transaction(func(tx *gorm.DB) error {
		var blogPost models.BlogPost
		if err := tx.First(&blogPost, ctx.Param("id")).Error; err != nil {
			return err
		}

		if err := tx.Delete(&blogPost).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		r.HandleError(ctx, "Failed to delete post", nil, err)
		return
	}

	r.HandleAdminPosts(ctx)
}

func (r *Router) HandleAdminPosts(ctx *gin.Context) {
	var posts []models.BlogPost
	if err := r.Db.Preload("Tags").Order("created_at DESC").Limit(10).Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.HTML(200, "index", addGenerics(ctx, gin.H{
		"title":     "Posts",
		"posts":     posts,
		"noPosts":   len(posts) == 0,
		"canDelete": true,
	}))
}

func (r *Router) HandleError(ctx *gin.Context, message string, fn func(ctx *gin.Context), err error) {
	toastId := uuid.New().String()

	hxRequest, exists := ctx.Get("isHXRequest")
	if exists && hxRequest.(bool) {
		ctx.Header("HX-Retarget", "#toastContainer")
		ctx.Header("HX-Reswap", "beforeend")
		ctx.HTML(200, "toast", gin.H{
			"toastId": "toast-" + toastId,
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

func addGenerics(ctx *gin.Context, h gin.H) gin.H {
	hxRequest, exists := ctx.Get("isHXRequest")
	h["isHXRequest"] = exists && hxRequest.(bool)
	h["adminRoute"] = adminRoute
	h["authToken"], h["authed"] = ctx.Get("authToken")
	var ok bool
	h["theme"], ok = ctx.Get("theme")
	if !ok {
		h["theme"] = "dark"
	}
	return h
}

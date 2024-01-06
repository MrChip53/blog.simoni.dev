package server

import (
	"blog.simoni.dev/auth"
	"blog.simoni.dev/models"
	"blog.simoni.dev/templates/admin"
	"blog.simoni.dev/templates/components"
	"blog.simoni.dev/templates/pages"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Router struct {
	Db *gorm.DB
}

func NewRouter(db *gorm.DB) *Router {
	return &Router{Db: db}
}

func (r *Router) HandlePasswordChange(ctx *gin.Context) {
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")

	if len(newPassword) < 8 {
		r.HandleError(ctx, "Password must be at least 8 characters", nil, nil)
		return
	}

	var uId interface{}
	var ok bool
	if uId, ok = ctx.Get("userId"); !ok {
		r.HandleError(ctx, "You must be logged in to change your password", nil, nil)
		return
	}
	var userId uint
	if userId, ok = uId.(uint); !ok {
		r.HandleError(ctx, "You must be logged in to change your password", nil, nil)
		return
	}
	var user models.User
	if err := r.Db.Where("id = ?", userId).First(&user).Error; err != nil {
		r.HandleError(ctx, "Failed to find user", nil, err)
		return
	}
	if match, err := user.VerifyPassword(oldPassword); err != nil || !match {
		r.HandleError(ctx, "Failed to change password", nil, err)
		return
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		r.HandleError(ctx, "Failed to change password", nil, err)
		return
	}

	if err := r.Db.Model(&models.User{}).Where("id = ?", userId).Update("password", hash).Error; err != nil {
		r.HandleError(ctx, "Failed to change password", nil, err)
		return
	}

	// TODO change redirect location
	ctx.Redirect(http.StatusFound, "/admin")
}

func (r *Router) HandleUsernameChange(ctx *gin.Context) {
	newUsername := ctx.PostForm("username")
	if len(newUsername) == 0 {
		r.HandleError(ctx, "Username cannot be empty", nil, nil)
		return
	}
	var uId interface{}
	var ok bool
	if uId, ok = ctx.Get("userId"); !ok {
		r.HandleError(ctx, "You must be logged in to change your username", nil, nil)
		return
	}
	var userId uint
	if userId, ok = uId.(uint); !ok {
		r.HandleError(ctx, "You must be logged in to change your username", nil, nil)
		return
	}
	err := r.Db.Model(&models.User{}).Where("id = ?", userId).Update("username", newUsername).Error
	if err != nil {
		r.HandleError(ctx, "Failed to update username", nil, err)
		return
	}
	var user models.User
	if err = r.Db.Where("id = ?", userId).First(&user).Error; err != nil {
		r.HandleError(ctx, "Failed to load new username.", nil, err)
		return
	}
	_, err = user.NewAuthTokens(ctx)
	if err != nil {
		r.HandleError(ctx, "Failed to generate tokens", nil, err)
		return
	}

	// TODO change redirect location
	ctx.Redirect(http.StatusFound, "/admin")
}

func (r *Router) HandleIndex(ctx *gin.Context) {
	var title = "mrchip53's blog"
	var posts []models.BlogPost
	if err := r.Db.Preload("Tags").Where("draft = false").Order("created_at DESC").Limit(10).Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
	indexHtml := pages.IndexPage(posts, false)
	indexHtml.Render(createContext(ctx, title), ctx.Writer)
}

func (r *Router) HandleSettings(ctx *gin.Context) {
	ctx.Header("HX-Theme", "retro")

	ctx.Header("HX-Retarget", "#toastContainer")
	ctx.Header("HX-Reswap", "beforeend")
	ctx.Status(http.StatusOK)
	toast := components.ToastComponent("toast-3824892389", "Switch theme")
	toast.Render(context.TODO(), ctx.Writer)
}

func (r *Router) HandleUser(ctx *gin.Context) {
	username := ctx.Param("username")
	var posts []models.BlogPost
	if err := r.Db.Preload("Tags").Where("author = ? AND draft = false", username).Order("created_at DESC").Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
	indexHtml := pages.IndexPage(posts, false)
	indexHtml.Render(createContext(ctx, username+"'s Page"), ctx.Writer)
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

	ctx.Status(http.StatusOK)
	html := components.CommentsComponent(comments)
	html.Render(context.TODO(), ctx.Writer)
}

func (r *Router) HandlePost(ctx *gin.Context) {
	month := ctx.Param("month")
	day := ctx.Param("day")
	year := ctx.Param("year")
	slug := ctx.Param("slug")

	var post models.BlogPost
	if err := r.Db.Preload("Tags").Where("day(published_at) = ? AND month(published_at) = ? AND year(published_at) = ? AND slug = ?", day, month, year, slug).First(&post).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		r.HandleNotFound(ctx)
		return
	}

	var comments []models.Comment
	r.Db.Where("blog_post_id = ?", post.ID).Order("created_at DESC").Find(&comments)

	postHtml := parseMarkdown([]byte(post.Content))

	ctx.Status(200)
	indexHtml := pages.PostPage(post, string(postHtml), comments)
	indexHtml.Render(createContext(ctx, post.Title), ctx.Writer)
}

func (r *Router) HandlePostEdit(ctx *gin.Context) {
	postId := ctx.Param("postId")

	var post models.BlogPost
	if err := r.Db.Preload("Tags").Where("id = ?", postId).First(&post).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		r.HandleNotFound(ctx)
		return
	}

	postHtml := parseMarkdown([]byte(post.Content))

	ctx.Status(http.StatusOK)
	html := admin.EditPostPage(post, string(postHtml))
	html.Render(createContext(ctx, "Editing "+post.Title), ctx.Writer)
}

func (r *Router) PostPostEdit(ctx *gin.Context) {
	postId := ctx.Param("postId")

	content := ctx.PostForm("content")
	publish := ctx.PostForm("publish")

	if len(content) == 0 {
		r.HandleError(ctx, "Just delete the post instead.", nil, nil)
		return
	}

	var post models.BlogPost
	if err := r.Db.Where("id = ?", postId).First(&post).Error; err != nil {
		r.HandleError(ctx, "Failed to load new content.", nil, err)
		return
	}

	post.Content = strings.TrimSpace(content)
	var wasPublished = !post.Draft
	post.Draft = publish != "on"
	if !wasPublished {
		post.Publish()
		post.Slug = url.QueryEscape(strings.ToLower(strings.ReplaceAll(post.Title, " ", "-")))
	}

	err := r.Db.Save(&post).Error
	if err != nil {
		r.HandleError(ctx, "Failed to update post.", nil, err)
		return
	}

	var location = adminRoute
	if !post.Draft {
		location = fmt.Sprintf("/post/%d/%d/%d/%s", post.CreatedAt.Month(), post.CreatedAt.Day(), post.CreatedAt.Year(), post.Slug)
	}
	ctx.Redirect(http.StatusFound, location)
	//ctx.Header("HX-Location", location)
}

func (r *Router) HandleLogin(ctx *gin.Context) {
	redirect := ctx.Request.URL.Query().Get("redirect")

	if redirect == "" {
		redirect = "/"
	}

	ctx.Status(http.StatusOK)

	html := pages.LoginPage(redirect, "")
	html.Render(createContext(ctx, "Login"), ctx.Writer)
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
	if err := r.Db.Preload("Tags").Where("id IN ? AND draft = false", postIds).Order("created_at DESC").Find(&posts).Error; err != nil {
		log.Println("Tag failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	indexHtml := pages.IndexPage(posts, false)
	indexHtml.Render(createContext(ctx, "Posts tagged with "+tag), ctx.Writer)
}

func (r *Router) HandleNotFound(ctx *gin.Context) {
	ctx.Status(http.StatusNotFound)
	html := pages.NotFoundPage()
	html.Render(createContext(ctx, "Oops!"), ctx.Writer)
}

func (r *Router) HandleInternalServerError(ctx *gin.Context) {
	ctx.Status(http.StatusInternalServerError)
	html := pages.NotFoundPage()
	html.Render(createContext(ctx, "Internal Server Error"), ctx.Writer)
}

func (r *Router) HandleHealth(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status": "ok",
	})
}

func (r *Router) HandleLogoutRequest(ctx *gin.Context) {
	auth.DeleteAuthCookies(ctx)
	ctx.Redirect(http.StatusFound, "/")
}

func (r *Router) HandleLoginRequest(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	redirectPath := ctx.PostForm("redirect")

	var errString string

	var user models.User
	if err := r.Db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Println("Login failed to get user:", err)
		errString = "Invalid username or password"
		r.HandleError(ctx, errString, func(ctx *gin.Context) {
			html := pages.LoginPage(redirectPath, errString)
			html.Render(createContext(ctx, "Login"), ctx.Writer)
		}, err)
		return
	}

	if match, err := user.VerifyPassword(password); err != nil {
		log.Println("Login failed to verify password:", err)
		errString = "Invalid username or password"
		r.HandleError(ctx, errString, func(ctx *gin.Context) {
			html := pages.LoginPage(redirectPath, errString)
			html.Render(createContext(ctx, "Login"), ctx.Writer)
		}, err)
		return
	} else if !match {
		errString = "Invalid username or password"
		r.HandleError(ctx, errString, func(ctx *gin.Context) {
			html := pages.LoginPage(redirectPath, errString)
			html.Render(createContext(ctx, "Login"), ctx.Writer)
		}, err)
		return
	}

	_, err := user.NewAuthTokens(ctx)
	if err != nil {
		log.Println("Login failed to generate tokens:", err)
		errString = "Invalid username or password"
		r.HandleError(ctx, errString, func(ctx *gin.Context) {
			html := pages.LoginPage(redirectPath, errString)
			html.Render(createContext(ctx, "Login"), ctx.Writer)
		}, err)
		return
	}
	//AddJwtPayloadToCtx(ctx, payload)
	ctx.Redirect(http.StatusFound, redirectPath)
	//html := components.Navbar(true)
	//html.Render(createContext(ctx, ""), ctx.Writer)
}

func (r *Router) HandleAdminDashboard(ctx *gin.Context) {
	claims := ctx.MustGet("authToken").(*auth.JwtPayload)
	if claims == nil {
		ctx.Redirect(http.StatusFound, "/login?redirect="+ctx.Request.URL.Path)
		ctx.Abort()
		return
	}

	var posts []models.BlogPost
	if err := r.Db.Preload("Tags").Where("draft = true").Order("created_at DESC").Limit(10).Find(&posts).Error; err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	html := admin.DashboardPage(posts, strconv.Itoa(len(posts)), "0")
	html.Render(createContext(ctx, "Admin Dashboard"), ctx.Writer)
}

func (r *Router) HandleAdminAddTagToPost(ctx *gin.Context) {
	postId := ctx.Param("id")
	tag := ctx.PostForm("tag")

	err := r.Db.Transaction(func(tx *gorm.DB) error {
		var blogPost models.BlogPost
		if err := tx.First(&blogPost, postId).Error; err != nil {
			return err
		}

		var tagModel models.Tag
		if err := tx.Where("name = ?", tag).First(&tagModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tagM, err := models.NewTag(tx, tag)
				if err != nil {
					return err
				}
				tagModel = *tagM
			} else {
				return err
			}
		}

		if err := tx.Model(&blogPost).Association("Tags").Append(&tagModel); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		r.HandleError(ctx, "Failed to add tag", nil, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (r *Router) HandleAdminDeleteTagFromPost(ctx *gin.Context) {
	postId := ctx.Param("id")
	tagId := ctx.Param("tagId")

	err := r.Db.Transaction(func(tx *gorm.DB) error {
		var blogPost models.BlogPost
		if err := tx.First(&blogPost, postId).Error; err != nil {
			return err
		}

		var tag models.Tag
		if err := tx.First(&tag, tagId).Error; err != nil {
			return err
		}

		if err := tx.Model(&blogPost).Association("Tags").Delete(&tag); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		r.HandleError(ctx, "Failed to delete tag", nil, err)
		return
	}

	ctx.Status(http.StatusOK)
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

	html := pages.LoginPage(redirect, "")
	html.Render(createContext(ctx, "Login"), ctx.Writer)
}

func (r *Router) HandleAdminNewBlogPost(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
	html := admin.NewPostPage()
	html.Render(createContext(ctx, "New Blog Post"), ctx.Writer)
}

func (r *Router) HandleAdminNewBlogPostRequest(ctx *gin.Context) {
	title := ctx.PostForm("title")
	tags := strings.Split(ctx.PostForm("tags"), ",")
	content := ctx.PostForm("content")
	description := ctx.PostForm("description")
	draft := ctx.PostForm("publish") != "true"
	slug := url.QueryEscape(strings.ToLower(strings.ReplaceAll(title, " ", "-")))

	jwt, exists := ctx.Get("authToken")
	if !exists {
		r.HandleError(ctx, "You are not logged in", nil, nil)
		return
	}
	author := jwt.(*auth.JwtPayload).Username

	tx := r.Db.Begin()
	newPost, err := models.NewBlogPost(tx, title, author, slug, strings.TrimSpace(content), description, draft)
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

	ctx.Status(200)
	indexHtml := pages.IndexPage(posts, true)
	indexHtml.Render(createContext(ctx, "Manage Posts"), ctx.Writer)
}

func (r *Router) HandleError(ctx *gin.Context, message string, fn func(ctx *gin.Context), err error) {
	toastId := uuid.New().String()

	hxRequest, exists := ctx.Get("isHXRequest")
	if exists && hxRequest.(bool) {
		ctx.Header("HX-Retarget", "#toastContainer")
		ctx.Header("HX-Reswap", "beforeend")
		ctx.Status(http.StatusOK)
		toast := components.ToastComponent("toast-"+toastId, message)
		toast.Render(context.TODO(), ctx.Writer)
		return
	}

	if fn != nil {
		fn(ctx)
	} else {
		r.HandleInternalServerError(ctx)
	}
}

func createContext(ctx *gin.Context, pageTitle string) context.Context {
	username, uOk := ctx.Get("username")
	theme, ok := ctx.Get("theme")
	_, aOk := ctx.Get("authed")
	isAdmin, _ := ctx.Get("isAdmin")
	hxRequest, exists := ctx.Get("isHXRequest")
	userId, _ := ctx.Get("userId")

	ct := context.WithValue(context.Background(), "isHxRequest", exists && hxRequest.(bool))
	ct = context.WithValue(ct, "adminRoute", adminRoute)
	if username != nil {
		ct = context.WithValue(ct, "username", username.(string))
	}
	ct = context.WithValue(ct, "authed", aOk)
	if isAdmin != nil {
		ct = context.WithValue(ct, "isAdmin", isAdmin.(bool))
	}
	if userId != nil {
		ct = context.WithValue(ct, "userId", userId.(uint))
	}
	if uOk {
		ct = context.WithValue(ct, "initials", username.(string)[:2])
	}
	if !ok {
		ct = context.WithValue(ct, "theme", "dark")
	} else {
		ct = context.WithValue(ct, "theme", theme.(string))
	}
	ct = context.WithValue(ct, "pageTitle", pageTitle)

	return ct
}

func addGenerics(ctx *gin.Context, h gin.H) gin.H {
	hxRequest, exists := ctx.Get("isHXRequest")
	h["isHXRequest"] = exists && hxRequest.(bool)
	h["adminRoute"] = adminRoute
	_, h["authed"] = ctx.Get("authed")
	h["isAdmin"], _ = ctx.Get("isAdmin")
	h["userId"], _ = ctx.Get("userId")
	var ok, uOk bool
	h["username"], uOk = ctx.Get("username")
	if uOk {
		h["initials"] = strings.ToUpper(h["username"].(string)[:2])
	}
	h["theme"], ok = ctx.Get("theme")
	if !ok {
		h["theme"] = "dark"
	}
	return h
}

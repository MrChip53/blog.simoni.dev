package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"blog.simoni.dev/auth"
	db "blog.simoni.dev/db/generated"
	"blog.simoni.dev/templates/admin"
	"blog.simoni.dev/templates/components"
	"blog.simoni.dev/templates/pages"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Router struct {
	Queries *db.Queries
	Pool    *pgxpool.Pool
}

func NewRouter(pool *pgxpool.Pool) *Router {
	queries := db.New(pool)
	return &Router{Pool: pool, Queries: queries}
}

func (r *Router) HandlePasswordChange(ctx *gin.Context) {
	oldPassword := ctx.PostForm("oldPassword")
	newPassword := ctx.PostForm("newPassword")

	if len(newPassword) < 8 {
		r.HandleError(ctx, "Password must be at least 8 characters", nil, nil)
		return
	}

	uId, ok := ctx.Get("userId")
	if !ok {
		r.HandleError(ctx, "You must be logged in to change your password", nil, nil)
		return
	}
	userId, ok := uId.(uint)
	if !ok {
		r.HandleError(ctx, "You must be logged in to change your password", nil, nil)
		return
	}

	row, err := r.Queries.GetUserByID(ctx.Request.Context(), int64(userId))
	if err != nil {
		r.HandleError(ctx, "Failed to find user", nil, err)
		return
	}
	user := mapUser(row)

	if match, err := user.VerifyPassword(oldPassword); err != nil || !match {
		r.HandleError(ctx, "Failed to change password", nil, err)
		return
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		r.HandleError(ctx, "Failed to change password", nil, err)
		return
	}

	if err := r.Queries.UpdateUserPassword(ctx.Request.Context(), db.UpdateUserPasswordParams{
		ID:       int64(userId),
		Password: hash,
	}); err != nil {
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
	uId, ok := ctx.Get("userId")
	if !ok {
		r.HandleError(ctx, "You must be logged in to change your username", nil, nil)
		return
	}
	userId, ok := uId.(uint)
	if !ok {
		r.HandleError(ctx, "You must be logged in to change your username", nil, nil)
		return
	}

	if err := r.Queries.UpdateUserUsername(ctx.Request.Context(), db.UpdateUserUsernameParams{
		ID:       int64(userId),
		Username: newUsername,
	}); err != nil {
		r.HandleError(ctx, "Failed to update username", nil, err)
		return
	}

	row, err := r.Queries.GetUserByID(ctx.Request.Context(), int64(userId))
	if err != nil {
		r.HandleError(ctx, "Failed to load new username.", nil, err)
		return
	}
	user := mapUser(row)

	if _, err = user.NewAuthTokens(ctx); err != nil {
		r.HandleError(ctx, "Failed to generate tokens", nil, err)
		return
	}

	// TODO change redirect location
	ctx.Redirect(http.StatusFound, "/admin")
}

func (r *Router) HandleIndex(ctx *gin.Context) {
	rows, err := r.Queries.GetPublishedPosts(ctx.Request.Context())
	if err != nil {
		log.Println("Index failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	posts, err := r.loadPostsWithTags(ctx.Request.Context(), rows)
	if err != nil {
		log.Println("Index failed to load tags:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
	pages.IndexPage(posts, false).Render(createContext(ctx, "mrchip53's blog"), ctx.Writer)
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
	rows, err := r.Queries.GetPostsByAuthor(ctx.Request.Context(), username)
	if err != nil {
		log.Println("User page failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	posts, err := r.loadPostsWithTags(ctx.Request.Context(), rows)
	if err != nil {
		log.Println("User page failed to load tags:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
	pages.IndexPage(posts, false).Render(createContext(ctx, username+"'s Page"), ctx.Writer)
}

func (r *Router) HandleComment(ctx *gin.Context) {
	if t, ok := ctx.Get("authToken"); !ok || t == nil {
		r.HandleError(ctx, "You must be logged in to comment", nil, nil)
		return
	}

	postIdStr := ctx.Param("postId")
	author := ctx.PostForm("Username")
	comment := ctx.PostForm("comment")

	if len(comment) == 0 || len(author) == 0 {
		r.HandleError(ctx, "Author and comment cannot be empty", nil, nil)
		return
	}

	pid, err := strconv.ParseInt(postIdStr, 10, 64)
	if err != nil {
		r.HandleError(ctx, "Invalid post ID", nil, err)
		return
	}

	if _, err := r.Queries.GetPostByID(ctx.Request.Context(), pid); err != nil {
		r.HandleError(ctx, "Post not found", nil, err)
		return
	}

	if _, err := r.Queries.CreateComment(ctx.Request.Context(), db.CreateCommentParams{
		BlogPostID: pid,
		Author:     author,
		Comment:    comment,
	}); err != nil {
		r.HandleError(ctx, "Failed to create comment", nil, err)
		return
	}

	dbComments, err := r.Queries.GetCommentsByPostID(ctx.Request.Context(), pid)
	if err != nil {
		r.HandleError(ctx, "Failed to load comments", nil, err)
		return
	}

	ctx.Header("HX-Retarget", "#comments-"+postIdStr)
	ctx.Header("HX-Reswap", "innerHTML")
	ctx.Status(http.StatusOK)
	components.CommentsComponent(mapComments(dbComments)).Render(context.TODO(), ctx.Writer)
}

func (r *Router) HandlePost(ctx *gin.Context) {
	month := ctx.Param("month")
	day := ctx.Param("day")
	year := ctx.Param("year")
	slug := ctx.Param("slug")

	m, _ := strconv.Atoi(month)
	d, _ := strconv.Atoi(day)
	y, _ := strconv.Atoi(year)
	// Use calendar-based next-day midnight so DST days (25h or 23h) are handled correctly.
	startOfDay := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local)
	endOfDay := time.Date(y, time.Month(m), d+1, 0, 0, 0, 0, time.Local)

	row, err := r.Queries.GetPostBySlugAndDate(ctx.Request.Context(), db.GetPostBySlugAndDateParams{
		Slug:       slug,
		StartOfDay: pgtype.Timestamptz{Time: startOfDay, Valid: true},
		EndOfDay:   pgtype.Timestamptz{Time: endOfDay, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.HandleNotFound(ctx)
		} else {
			log.Println("Post page failed:", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	dbTags, _ := r.Queries.GetTagsForPost(ctx.Request.Context(), row.ID)
	post := mapPost(row, mapTags(dbTags))

	dbComments, _ := r.Queries.GetCommentsByPostID(ctx.Request.Context(), row.ID)

	ctx.Status(200)
	pages.PostPage(post, string(parseMarkdown([]byte(post.Content))), mapComments(dbComments)).Render(createContext(ctx, post.Title), ctx.Writer)
}

func (r *Router) HandlePostEdit(ctx *gin.Context) {
	postId, err := strconv.ParseInt(ctx.Param("postId"), 10, 64)
	if err != nil {
		r.HandleNotFound(ctx)
		return
	}

	row, err := r.Queries.GetPostByID(ctx.Request.Context(), postId)
	if err != nil {
		log.Println("PostEdit failed to get post:", err)
		r.HandleNotFound(ctx)
		return
	}

	dbTags, _ := r.Queries.GetTagsForPost(ctx.Request.Context(), postId)
	post := mapPost(row, mapTags(dbTags))

	ctx.Status(http.StatusOK)
	admin.EditPostPage(post, string(parseMarkdown([]byte(post.Content)))).Render(createContext(ctx, "Editing "+post.Title), ctx.Writer)
}

func (r *Router) PostPostEdit(ctx *gin.Context) {
	postId, err := strconv.ParseInt(ctx.Param("postId"), 10, 64)
	if err != nil {
		r.HandleError(ctx, "Invalid post ID", nil, err)
		return
	}

	content := ctx.PostForm("content")
	publish := ctx.PostForm("publish")

	if len(content) == 0 {
		r.HandleError(ctx, "Just delete the post instead.", nil, nil)
		return
	}

	row, err := r.Queries.GetPostByID(ctx.Request.Context(), postId)
	if err != nil {
		r.HandleError(ctx, "Failed to load post.", nil, err)
		return
	}

	draft := publish != "on"
	publishedAt := row.PublishedAt
	slug := row.Slug
	if row.Draft && !draft {
		// First time publishing
		publishedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
		slug = url.QueryEscape(strings.ToLower(strings.ReplaceAll(row.Title, " ", "-")))
	}

	updated, err := r.Queries.UpdatePost(ctx.Request.Context(), db.UpdatePostParams{
		ID:          postId,
		Title:       row.Title,
		Content:     strings.TrimSpace(content),
		Slug:        slug,
		Draft:       draft,
		PublishedAt: publishedAt,
	})
	if err != nil {
		r.HandleError(ctx, "Failed to update post.", nil, err)
		return
	}

	location := adminRoute
	if !updated.Draft {
		if t := pgTimeToTime(updated.PublishedAt); !t.IsZero() {
			tl := t.Local()
			location = fmt.Sprintf("/post/%02d/%02d/%d/%s", tl.Month(), tl.Day(), tl.Year(), updated.Slug)
		}
	}
	ctx.Redirect(http.StatusFound, location)
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

	rows, err := r.Queries.GetPublishedPostsByTag(ctx.Request.Context(), tag)
	if err != nil {
		log.Println("Tag failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	posts, err := r.loadPostsWithTags(ctx.Request.Context(), rows)
	if err != nil {
		log.Println("Tag failed to load tags:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	pages.IndexPage(posts, false).Render(createContext(ctx, "Posts tagged with "+tag), ctx.Writer)
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

	errString := "Invalid username or password"

	row, err := r.Queries.GetUserByUsername(ctx.Request.Context(), username)
	if err != nil {
		time.Sleep(time.Duration(170+rand.Intn(35)) * time.Millisecond)
		log.Println("Login failed to get user:", err)
		r.HandleError(ctx, errString, func(ctx *gin.Context) {
			pages.LoginPage(redirectPath, errString).Render(createContext(ctx, "Login"), ctx.Writer)
		}, err)
		return
	}
	user := mapUser(row)

	match, err := user.VerifyPassword(password)
	if err != nil || !match {
		if err != nil {
			log.Println("Login failed to verify password:", err)
		}
		r.HandleError(ctx, errString, func(ctx *gin.Context) {
			pages.LoginPage(redirectPath, errString).Render(createContext(ctx, "Login"), ctx.Writer)
		}, err)
		return
	}

	if _, err := user.NewAuthTokens(ctx); err != nil {
		log.Println("Login failed to generate tokens:", err)
		r.HandleError(ctx, errString, func(ctx *gin.Context) {
			pages.LoginPage(redirectPath, errString).Render(createContext(ctx, "Login"), ctx.Writer)
		}, err)
		return
	}

	ctx.Redirect(http.StatusFound, redirectPath)
}

func (r *Router) HandleAdminDashboard(ctx *gin.Context) {
	claims := ctx.MustGet("authToken").(*auth.JwtPayload)
	if claims == nil {
		ctx.Redirect(http.StatusFound, "/login?redirect="+ctx.Request.URL.Path)
		ctx.Abort()
		return
	}

	rows, err := r.Queries.GetDraftPosts(ctx.Request.Context())
	if err != nil {
		log.Println("Dashboard failed to get posts:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	posts, err := r.loadPostsWithTags(ctx.Request.Context(), rows)
	if err != nil {
		log.Println("Dashboard failed to load tags:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	admin.DashboardPage(posts, strconv.Itoa(len(posts)), "0").Render(createContext(ctx, "Admin Dashboard"), ctx.Writer)
}

func (r *Router) HandleAdminAddTagToPost(ctx *gin.Context) {
	postId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		r.HandleError(ctx, "Invalid post ID", nil, err)
		return
	}
	tag := ctx.PostForm("tag")

	tx, err := r.Pool.Begin(ctx.Request.Context())
	if err != nil {
		r.HandleError(ctx, "Failed to add tag", nil, err)
		return
	}
	defer tx.Rollback(ctx.Request.Context())
	qtx := r.Queries.WithTx(tx)

	tagRow, err := qtx.GetTagByName(ctx.Request.Context(), tag)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			r.HandleError(ctx, "Failed to add tag", nil, err)
			return
		}
		tagRow, err = qtx.CreateTag(ctx.Request.Context(), tag)
		if err != nil {
			r.HandleError(ctx, "Failed to create tag", nil, err)
			return
		}
	}

	if err := qtx.AddTagToPost(ctx.Request.Context(), db.AddTagToPostParams{
		BlogPostID: postId,
		TagID:      tagRow.ID,
	}); err != nil {
		r.HandleError(ctx, "Failed to add tag", nil, err)
		return
	}

	if err := tx.Commit(ctx.Request.Context()); err != nil {
		r.HandleError(ctx, "Failed to add tag", nil, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (r *Router) HandleAdminDeleteTagFromPost(ctx *gin.Context) {
	postId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		r.HandleError(ctx, "Invalid post ID", nil, err)
		return
	}
	tagId, err := strconv.ParseInt(ctx.Param("tagId"), 10, 64)
	if err != nil {
		r.HandleError(ctx, "Invalid tag ID", nil, err)
		return
	}

	if err := r.Queries.RemoveTagFromPost(ctx.Request.Context(), db.RemoveTagFromPostParams{
		BlogPostID: postId,
		TagID:      tagId,
	}); err != nil {
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

	var publishedAt pgtype.Timestamptz
	if !draft {
		publishedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	}

	tx, err := r.Pool.Begin(ctx.Request.Context())
	if err != nil {
		r.HandleError(ctx, "Failed to create blog post", nil, err)
		return
	}
	defer tx.Rollback(ctx.Request.Context())
	qtx := r.Queries.WithTx(tx)

	post, err := qtx.CreatePost(ctx.Request.Context(), db.CreatePostParams{
		Title:       title,
		Author:      author,
		Slug:        slug,
		Content:     strings.TrimSpace(content),
		Description: description,
		Draft:       draft,
		PublishedAt: publishedAt,
	})
	if err != nil {
		r.HandleError(ctx, "Failed to create blog post", nil, err)
		return
	}

	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		tagRow, err := qtx.GetTagByName(ctx.Request.Context(), tagName)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				r.HandleError(ctx, "Failed to create blog post", nil, err)
				return
			}
			tagRow, err = qtx.CreateTag(ctx.Request.Context(), tagName)
			if err != nil {
				r.HandleError(ctx, "Failed to create blog post", nil, err)
				return
			}
		}
		if err := qtx.AddTagToPost(ctx.Request.Context(), db.AddTagToPostParams{
			BlogPostID: post.ID,
			TagID:      tagRow.ID,
		}); err != nil {
			r.HandleError(ctx, "Failed to create blog post", nil, err)
			return
		}
	}

	if err := tx.Commit(ctx.Request.Context()); err != nil {
		r.HandleError(ctx, "Failed to create blog post", nil, err)
		return
	}

	ctx.Redirect(302, adminRoute)
}

func (r *Router) HandleAdminPostsDelete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		r.HandleError(ctx, "Invalid post ID", nil, err)
		return
	}
	if err := r.Queries.SoftDeletePost(ctx.Request.Context(), id); err != nil {
		r.HandleError(ctx, "Failed to delete post", nil, err)
		return
	}
	r.HandleAdminPosts(ctx)
}

func (r *Router) HandleAdminPosts(ctx *gin.Context) {
	rows, err := r.Queries.GetAllPostsAdmin(ctx.Request.Context())
	if err != nil {
		log.Println("Admin posts failed:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	posts, err := r.loadPostsWithTags(ctx.Request.Context(), rows)
	if err != nil {
		log.Println("Admin posts failed to load tags:", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.Status(200)
	pages.IndexPage(posts, true).Render(createContext(ctx, "Manage Posts"), ctx.Writer)
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

func (r *Router) HandleWasmLoader(ctx *gin.Context) {
	wasmType := ctx.Param("type")
	url := ctx.Query("url")
	ctx.Status(http.StatusOK)
	html := pages.WasmPage(wasmType, url)
	html.Render(createContext(ctx, "Wasm Loader"), ctx.Writer)
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

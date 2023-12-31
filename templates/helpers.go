package templates

import (
	"blog.simoni.dev/models"
	"context"
	"fmt"
	"github.com/a-h/templ"
	"time"
)

func GetThemeLink(ctx context.Context) string {
	theme, ok := ctx.Value("theme").(string)
	if !ok {
		return "/css/themes/dark.css"
	}
	return "/css/themes/" + theme + ".css"
}

func GetTagLink(tag string) templ.SafeURL {
	return templ.SafeURL("/tag/" + tag)
}

func GetPostSlug(post models.BlogPost) templ.SafeURL {
	return templ.SafeURL(fmt.Sprintf("/post/%02d/%02d/%d/%s", post.CreatedAt.Month(), post.CreatedAt.Day(), post.CreatedAt.Year(), post.Slug))
}

func GetUserLink(username string) templ.SafeURL {
	return templ.SafeURL("/user/" + username)
}

func GetDeletePostLink(adminRoute string, postId uint) string {
	return fmt.Sprintf("%s/post/%d", adminRoute, postId)
}

func GetTitle(ctx context.Context) string {
	title, ok := ctx.Value("title").(string)
	if !ok {
		return "mrchip53's blog"
	}
	return title
}

func GenerateContent(ctx context.Context) string {
	content, ok := ctx.Value("contentFunction").(func() string)
	if !ok {
		return ""
	}
	return content()
}

func GetAdminRoute(ctx context.Context) string {
	adminRoute, ok := ctx.Value("adminRoute").(string)
	if !ok {
		return "/admin"
	}
	return adminRoute
}

func isHxRequest(ctx context.Context) bool {
	hxRequest, ok := ctx.Value("isHxRequest").(bool)
	if !ok {
		return false
	}
	return hxRequest
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

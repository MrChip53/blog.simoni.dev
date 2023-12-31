package server

import (
	"blog.simoni.dev/md"
	"blog.simoni.dev/models"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"time"
)

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

func parseMarkdown(bytes []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(bytes)

	renderer := md.NewRenderer()

	return markdown.Render(doc, renderer)
}

func getSlug(post models.BlogPost) string {
	return fmt.Sprintf("/post/%02d/%02d/%d/%s", post.CreatedAt.Month(), post.CreatedAt.Day(), post.CreatedAt.Year(), post.Slug)
}

func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}

	return s
}

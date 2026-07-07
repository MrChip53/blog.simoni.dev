package server

import (
	"context"
	"time"

	db "blog.simoni.dev/db/generated"
	"blog.simoni.dev/models"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgTimeToTime(t pgtype.Timestamptz) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func pgTimeToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func mapTag(t db.Tag) models.Tag {
	return models.Tag{
		ID:        t.ID,
		CreatedAt: pgTimeToTime(t.CreatedAt),
		Name:      t.Name,
	}
}

func mapTags(tags []db.Tag) []models.Tag {
	result := make([]models.Tag, len(tags))
	for i, t := range tags {
		result[i] = mapTag(t)
	}
	return result
}

func mapPost(p db.BlogPost, tags []models.Tag) models.BlogPost {
	return models.BlogPost{
		ID:          p.ID,
		CreatedAt:   pgTimeToTime(p.CreatedAt),
		UpdatedAt:   pgTimeToTime(p.UpdatedAt),
		Title:       p.Title,
		Author:      p.Author,
		Slug:        p.Slug,
		Content:     p.Content,
		Description: p.Description,
		Draft:       p.Draft,
		PublishedAt: pgTimeToTimePtr(p.PublishedAt),
		Tags:        tags,
	}
}

func mapComment(c db.Comment) models.Comment {
	return models.Comment{
		ID:         c.ID,
		CreatedAt:  pgTimeToTime(c.CreatedAt),
		BlogPostId: c.BlogPostID,
		Author:     c.Author,
		Comment:    c.Comment,
	}
}

func mapComments(comments []db.Comment) []models.Comment {
	result := make([]models.Comment, len(comments))
	for i, c := range comments {
		result[i] = mapComment(c)
	}
	return result
}

func mapUser(u db.User) models.User {
	return models.User{
		ID:       u.ID,
		Username: u.Username,
		Password: u.Password,
		Admin:    u.Admin,
		Theme:    u.Theme,
	}
}

func (r *Router) loadPostsWithTags(ctx context.Context, posts []db.BlogPost) ([]models.BlogPost, error) {
	result := make([]models.BlogPost, len(posts))
	for i, p := range posts {
		tags, err := r.Queries.GetTagsForPost(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		result[i] = mapPost(p, mapTags(tags))
	}
	return result, nil
}

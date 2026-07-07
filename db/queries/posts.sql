-- name: GetPublishedPosts :many
SELECT * FROM blog_posts
WHERE draft = false AND deleted_at IS NULL
ORDER BY created_at DESC LIMIT 10;

-- name: GetPostBySlugAndDate :one
SELECT * FROM blog_posts
WHERE slug = @slug
  AND published_at >= @start_of_day
  AND published_at < @end_of_day
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetPostByID :one
SELECT * FROM blog_posts WHERE id = @id AND deleted_at IS NULL LIMIT 1;

-- name: CreatePost :one
INSERT INTO blog_posts (title, author, slug, content, description, draft, published_at)
VALUES (@title, @author, @slug, @content, @description, @draft, @published_at)
RETURNING *;

-- name: UpdatePost :one
UPDATE blog_posts
SET title = @title, content = @content, slug = @slug,
    draft = @draft, published_at = @published_at, updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePost :exec
UPDATE blog_posts SET deleted_at = NOW() WHERE id = @id;

-- name: GetPostsByAuthor :many
SELECT * FROM blog_posts
WHERE author = @author AND draft = false AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetAllPostsAdmin :many
SELECT * FROM blog_posts WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 10;

-- name: GetDraftPosts :many
SELECT * FROM blog_posts
WHERE draft = true AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 10;
-- name: CreateComment :one
INSERT INTO comments (blog_post_id, author, comment)
VALUES (@blog_post_id, @author, @comment)
RETURNING *;

-- name: GetCommentsByPostID :many
SELECT * FROM comments
WHERE blog_post_id = @blog_post_id AND deleted_at IS NULL
ORDER BY created_at DESC;
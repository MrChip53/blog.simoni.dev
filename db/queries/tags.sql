-- name: GetTagByName :one
SELECT * FROM tags WHERE name = @name AND deleted_at IS NULL LIMIT 1;

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = @id AND deleted_at IS NULL LIMIT 1;

-- name: CreateTag :one
INSERT INTO tags (name) VALUES (@name) RETURNING *;

-- name: GetTagsForPost :many
SELECT t.* FROM tags t
JOIN blog_post_tags bpt ON bpt.tag_id = t.id
WHERE bpt.blog_post_id = @post_id AND t.deleted_at IS NULL;

-- name: AddTagToPost :exec
INSERT INTO blog_post_tags (blog_post_id, tag_id)
VALUES (@blog_post_id, @tag_id)
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromPost :exec
DELETE FROM blog_post_tags WHERE blog_post_id = @blog_post_id AND tag_id = @tag_id;

-- name: GetPostIDsByTag :many
SELECT bpt.blog_post_id FROM blog_post_tags bpt
JOIN tags t ON t.id = bpt.tag_id
WHERE t.name = @name AND t.deleted_at IS NULL;

-- name: GetPublishedPostsByTag :many
SELECT * FROM blog_posts
WHERE id IN (
    SELECT bpt.blog_post_id FROM blog_post_tags bpt
    JOIN tags t ON t.id = bpt.tag_id
    WHERE t.name = @name AND t.deleted_at IS NULL
)
AND draft = false AND deleted_at IS NULL
ORDER BY created_at DESC;
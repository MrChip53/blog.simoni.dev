-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    username    VARCHAR(100) UNIQUE NOT NULL,
    password    VARCHAR(302) NOT NULL,
    admin       BOOLEAN NOT NULL DEFAULT FALSE,
    theme       VARCHAR(100) NOT NULL DEFAULT 'dark'
);

CREATE TABLE IF NOT EXISTS blog_posts (
    id           BIGSERIAL PRIMARY KEY,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    title        TEXT NOT NULL,
    author       TEXT NOT NULL,
    slug         VARCHAR(100) UNIQUE NOT NULL,
    content      TEXT NOT NULL,
    description  TEXT NOT NULL,
    draft        BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS tags (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    name       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS blog_post_tags (
    blog_post_id BIGINT NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    tag_id       BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (blog_post_id, tag_id)
);

CREATE TABLE IF NOT EXISTS comments (
    id           BIGSERIAL PRIMARY KEY,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,
    blog_post_id BIGINT NOT NULL,
    author       TEXT NOT NULL,
    comment      VARCHAR(1000) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS blog_post_tags;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS blog_posts;
DROP TABLE IF EXISTS users;